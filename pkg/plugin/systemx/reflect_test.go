package systemx

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewReflectNode(t *testing.T) {
	s := scheme.New()
	st, _ := storage.New(context.Background(), storage.Config{
		Scheme:   s,
		Database: memdb.New(faker.Word()),
	})

	n := NewReflectNode(ReflectNodeConfig{
		OP:      OPSelect,
		Storage: st,
	})
	assert.NotNil(t, n)
	assert.NotZero(t, n.ID())

	_ = n.Close()
}

func TestReflectNode_Send(t *testing.T) {
	s := scheme.New()

	st, _ := storage.New(context.Background(), storage.Config{
		Scheme:   s,
		Database: memdb.New(faker.Word()),
	})

	s.AddKnownType(KindReflect, &ReflectSpec{})
	s.AddCodec(KindReflect, scheme.CodecWithType[*ReflectSpec](func(spec *ReflectSpec) (node.Node, error) {
		return NewReflectNode(ReflectNodeConfig{
			ID:      spec.ID,
			OP:      spec.OP,
			Storage: st,
		}), nil
	}))

	t.Run(OPDelete, func(t *testing.T) {
		n := NewReflectNode(ReflectNodeConfig{
			OP:      OPDelete,
			Storage: st,
		})
		defer func() { _ = n.Close() }()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		id, _ := st.InsertOne(context.Background(), &ReflectSpec{
			SpecMeta: scheme.SpecMeta{
				ID:   ulid.Make(),
				Kind: KindReflect,
			},
			OP: OPDelete,
		})

		inPayload := primitive.NewMap(
			primitive.NewString(scheme.KeyID), primitive.NewString(id.String()),
		)
		inPck := packet.New(inPayload)

		ioStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioStream.Receive():
			outPayload, ok := outPck.Payload().(*primitive.Map)
			assert.True(t, ok)
			assert.Equal(t, id.String(), primitive.Interface(outPayload.GetOr(primitive.NewString(scheme.KeyID), nil)))
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run(OPInsert, func(t *testing.T) {
		n := NewReflectNode(ReflectNodeConfig{
			OP:      OPInsert,
			Storage: st,
		})
		defer func() { _ = n.Close() }()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		inPayload := primitive.NewMap(
			primitive.NewString(scheme.KeyID), primitive.NewString(ulid.Make().String()),
			primitive.NewString("kind"), primitive.NewString(KindReflect),
			primitive.NewString("op"), primitive.NewString(OPInsert),
		)
		inPck := packet.New(inPayload)

		ioStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioStream.Receive():
			outPayload, ok := outPck.Payload().(*primitive.Map)
			assert.True(t, ok)
			assert.NotNil(t, outPayload.GetOr(primitive.NewString(scheme.KeyID), nil))
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run(OPSelect, func(t *testing.T) {
		n := NewReflectNode(ReflectNodeConfig{
			OP:      OPSelect,
			Storage: st,
		})
		defer func() { _ = n.Close() }()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		id, _ := st.InsertOne(context.Background(), &ReflectSpec{
			SpecMeta: scheme.SpecMeta{
				ID:   ulid.Make(),
				Kind: KindReflect,
			},
			OP: OPSelect,
		})

		inPayload := primitive.NewMap(
			primitive.NewString(scheme.KeyID), primitive.NewString(id.String()),
		)
		inPck := packet.New(inPayload)

		ioStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioStream.Receive():
			outPayload, ok := outPck.Payload().(*primitive.Map)
			assert.True(t, ok)
			assert.Equal(t, id.String(), primitive.Interface(outPayload.GetOr(primitive.NewString(scheme.KeyID), nil)))
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run(OPUpdate, func(t *testing.T) {
		n := NewReflectNode(ReflectNodeConfig{
			OP:      OPUpdate,
			Storage: st,
		})
		defer func() { _ = n.Close() }()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		id, _ := st.InsertOne(context.Background(), &ReflectSpec{
			SpecMeta: scheme.SpecMeta{
				ID:   ulid.Make(),
				Kind: KindReflect,
			},
			OP: OPInsert,
		})

		inPayload := primitive.NewMap(
			primitive.NewString(scheme.KeyID), primitive.NewString(id.String()),
			primitive.NewString("op"), primitive.NewString(OPUpdate),
		)
		inPck := packet.New(inPayload)

		ioStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioStream.Receive():
			outPayload, ok := outPck.Payload().(*primitive.Map)
			assert.True(t, ok)
			assert.Equal(t, id.String(), primitive.Interface(outPayload.GetOr(primitive.NewString(scheme.KeyID), nil)))
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})
}

func BenchmarkReflectNode_Send(b *testing.B) {
	s := scheme.New()

	st, _ := storage.New(context.Background(), storage.Config{
		Scheme:   s,
		Database: memdb.New(faker.Word()),
	})

	s.AddKnownType(KindReflect, &ReflectSpec{})
	s.AddCodec(KindReflect, scheme.CodecWithType[*ReflectSpec](func(spec *ReflectSpec) (node.Node, error) {
		return NewReflectNode(ReflectNodeConfig{
			ID:      spec.ID,
			OP:      spec.OP,
			Storage: st,
		}), nil
	}))

	b.Run(OPDelete, func(b *testing.B) {
		n := NewReflectNode(ReflectNodeConfig{
			OP:      OPDelete,
			Storage: st,
		})
		defer func() { _ = n.Close() }()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			b.StopTimer()

			id, _ := st.InsertOne(context.Background(), &ReflectSpec{
				SpecMeta: scheme.SpecMeta{
					ID:   ulid.Make(),
					Kind: KindReflect,
				},
				OP: OPDelete,
			})

			inPayload := primitive.NewMap(
				primitive.NewString(scheme.KeyID), primitive.NewString(id.String()),
			)
			inPck := packet.New(inPayload)

			b.StartTimer()

			ioStream.Send(inPck)
			<-ioStream.Receive()
		}
	})

	b.Run(OPInsert, func(b *testing.B) {
		n := NewReflectNode(ReflectNodeConfig{
			OP:      OPInsert,
			Storage: st,
		})
		defer func() { _ = n.Close() }()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			b.StopTimer()

			inPayload := primitive.NewMap(
				primitive.NewString(scheme.KeyID), primitive.NewString(ulid.Make().String()),
				primitive.NewString("kind"), primitive.NewString(KindReflect),
				primitive.NewString("op"), primitive.NewString(OPInsert),
			)
			inPck := packet.New(inPayload)

			b.StartTimer()

			ioStream.Send(inPck)
			<-ioStream.Receive()
		}
	})

	b.Run(OPSelect, func(b *testing.B) {
		n := NewReflectNode(ReflectNodeConfig{
			OP:      OPSelect,
			Storage: st,
		})
		defer func() { _ = n.Close() }()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		id, _ := st.InsertOne(context.Background(), &ReflectSpec{
			SpecMeta: scheme.SpecMeta{
				ID:   ulid.Make(),
				Kind: KindReflect,
			},
			OP: OPSelect,
		})

		inPayload := primitive.NewMap(
			primitive.NewString(scheme.KeyID), primitive.NewString(id.String()),
		)
		inPck := packet.New(inPayload)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			ioStream.Send(inPck)
			<-ioStream.Receive()
		}
	})

	b.Run(OPUpdate, func(b *testing.B) {
		n := NewReflectNode(ReflectNodeConfig{
			OP:      OPUpdate,
			Storage: st,
		})
		defer func() { _ = n.Close() }()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			b.StopTimer()

			id, _ := st.InsertOne(context.Background(), &ReflectSpec{
				SpecMeta: scheme.SpecMeta{
					ID:   ulid.Make(),
					Kind: KindReflect,
				},
				OP: OPInsert,
			})

			inPayload := primitive.NewMap(
				primitive.NewString(scheme.KeyID), primitive.NewString(id.String()),
				primitive.NewString("op"), primitive.NewString(OPUpdate),
			)
			inPck := packet.New(inPayload)

			b.StartTimer()

			ioStream.Send(inPck)
			<-ioStream.Receive()
		}
	})
}
