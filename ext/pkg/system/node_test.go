package system

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestCreateNodes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	kind := faker.UUIDHyphenated()

	st, _ := store.New(ctx, memdb.NewCollection(""))

	n, _ := NewSyscallNode(CreateNodes(st))
	defer n.Close()

	meta := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload, _ := types.TextEncoder.Encode(meta)
	inPck := packet.New(types.NewSlice(inPayload))

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		var outPayload []*spec.Meta
		assert.NoError(t, types.Decoder.Decode(outPck.Payload(), &outPayload))
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func TestReadNodes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	kind := faker.UUIDHyphenated()

	st, _ := store.New(ctx, memdb.NewCollection(""))

	n, _ := NewSyscallNode(ReadNodes(st))
	defer n.Close()

	meta := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	id, _ := st.InsertOne(ctx, meta)

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload, _ := types.TextEncoder.Encode(store.Where[uuid.UUID]("id").EQ(id))
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		var outPayload []*spec.Meta
		assert.NoError(t, types.Decoder.Decode(outPck.Payload(), &outPayload))
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func TestUpdateNodes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	kind := faker.UUIDHyphenated()

	st, _ := store.New(ctx, memdb.NewCollection(""))

	n, _ := NewSyscallNode(UpdateNodes(st))
	defer n.Close()

	meta := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	_, _ = st.InsertOne(ctx, meta)

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload, _ := types.TextEncoder.Encode(meta)
	inPck := packet.New(types.NewSlice(inPayload))

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		var outPayload []*spec.Meta
		assert.NoError(t, types.Decoder.Decode(outPck.Payload(), &outPayload))
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func TestDeleteNodes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	kind := faker.UUIDHyphenated()

	st, _ := store.New(ctx, memdb.NewCollection(""))

	n, _ := NewSyscallNode(DeleteNodes(st))
	defer n.Close()

	meta := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	id, _ := st.InsertOne(ctx, meta)

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload, _ := types.TextEncoder.Encode(store.Where[uuid.UUID]("id").EQ(id))
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		var outPayload []*spec.Meta
		assert.NoError(t, types.Decoder.Decode(outPck.Payload(), &outPayload))
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}
