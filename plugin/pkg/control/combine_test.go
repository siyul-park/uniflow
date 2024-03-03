package control

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewCombineNode(t *testing.T) {
	n := NewCombineNode()
	assert.NotNil(t, n)
	assert.Equal(t, -1, n.Depth())
	assert.Equal(t, false, n.Inplace())

	assert.NoError(t, n.Close())
}

func TestCombineNode_SendAndReceive(t *testing.T) {
	t.Run("depth = 0", func(t *testing.T) {
		n := NewCombineNode()
		defer n.Close()

		n.SetDepth(0)

		var ins []*port.OutPort
		for i := 0; i < 4; i++ {
			in := port.NewOut()
			in.Link(n.In(node.MultiPort(node.PortIn, i)))
			ins = append(ins, in)
		}

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		proc := process.New()
		defer proc.Close()

		inWriters := make([]*port.Writer, len(ins))
		for i, in := range ins {
			inWriters[i] = in.Open(proc)
		}
		outReader := out.Open(proc)

		var inPayloads []primitive.Value
		for range inWriters {
			inPayloads = append(inPayloads, primitive.NewString(faker.UUIDHyphenated()))
		}

		combined := primitive.NewSlice(inPayloads...).Interface()

		for i, inWriter := range inWriters {
			inPck := packet.New(inPayloads[i])
			inWriter.Write(inPck)
		}

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-outReader.Read():
			assert.Equal(t, combined, outPck.Payload().Interface())
			outReader.Receive(outPck)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		for _, inWriter := range inWriters {
			select {
			case backPck := <-inWriter.Receive():
				assert.NotNil(t, backPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}
		}
	})

	t.Run("depth = 1", func(t *testing.T) {
		n := NewCombineNode()
		defer n.Close()

		n.SetDepth(1)

		var ins []*port.OutPort
		for i := 0; i < 4; i++ {
			in := port.NewOut()
			in.Link(n.In(node.MultiPort(node.PortIn, i)))
			ins = append(ins, in)
		}

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		proc := process.New()
		defer proc.Close()

		inWriters := make([]*port.Writer, len(ins))
		for i, in := range ins {
			inWriters[i] = in.Open(proc)
		}
		outReader := out.Open(proc)

		var inPayloads []primitive.Value
		for range inWriters {
			inPayloads = append(inPayloads, primitive.NewString(faker.UUIDHyphenated()))
		}

		combined := inPayloads[len(inPayloads)-1].Interface()

		for i, inWriter := range inWriters {
			inPck := packet.New(inPayloads[i])
			inWriter.Write(inPck)
		}

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-outReader.Read():
			assert.Equal(t, combined, outPck.Payload().Interface())
			outReader.Receive(outPck)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		for _, inWriter := range inWriters {
			select {
			case backPck := <-inWriter.Receive():
				assert.NotNil(t, backPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}
		}
	})

	t.Run("depth = 2", func(t *testing.T) {
		n := NewCombineNode()
		defer n.Close()

		n.SetDepth(2)

		var ins []*port.OutPort
		for i := 0; i < 4; i++ {
			in := port.NewOut()
			in.Link(n.In(node.MultiPort(node.PortIn, i)))
			ins = append(ins, in)
		}

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		proc := process.New()
		defer proc.Close()

		inWriters := make([]*port.Writer, len(ins))
		for i, in := range ins {
			inWriters[i] = in.Open(proc)
		}
		outReader := out.Open(proc)

		var inPayloads []primitive.Value
		combined := map[string]string{}
		for range inWriters {
			key := faker.UUIDHyphenated()
			value := faker.UUIDHyphenated()

			inPayloads = append(inPayloads, primitive.NewMap(primitive.NewString(key), primitive.NewString(value)))
			combined[key] = value
		}

		for i, inWriter := range inWriters {
			inPck := packet.New(inPayloads[i])
			inWriter.Write(inPck)
		}

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-outReader.Read():
			assert.Equal(t, combined, outPck.Payload().Interface())
			outReader.Receive(outPck)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		for _, inWriter := range inWriters {
			select {
			case backPck := <-inWriter.Receive():
				assert.NotNil(t, backPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}
		}
	})

	t.Run("depth = -1", func(t *testing.T) {
		n := NewCombineNode()
		defer n.Close()

		n.SetDepth(-1)

		var ins []*port.OutPort
		for i := 0; i < 4; i++ {
			in := port.NewOut()
			in.Link(n.In(node.MultiPort(node.PortIn, i)))
			ins = append(ins, in)
		}

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		proc := process.New()
		defer proc.Close()

		inWriters := make([]*port.Writer, len(ins))
		for i, in := range ins {
			inWriters[i] = in.Open(proc)
		}
		outReader := out.Open(proc)

		var inPayloads []primitive.Value
		var combined []map[string]string
		for range inWriters {
			key := faker.UUIDHyphenated()
			value := faker.UUIDHyphenated()

			inPayloads = append(inPayloads, primitive.NewSlice(primitive.NewMap(primitive.NewString(key), primitive.NewString(value))))
			combined = append(combined, map[string]string{key: value})
		}

		for i, inWriter := range inWriters {
			inPck := packet.New(inPayloads[i])
			inWriter.Write(inPck)
		}

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-outReader.Read():
			assert.Equal(t, combined, outPck.Payload().Interface())
			outReader.Receive(outPck)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		for _, inWriter := range inWriters {
			select {
			case backPck := <-inWriter.Receive():
				assert.NotNil(t, backPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}
		}
	})

	t.Run("inplace = true", func(t *testing.T) {
		n := NewCombineNode()
		defer n.Close()

		n.SetInplace(true)

		var ins []*port.OutPort
		for i := 0; i < 4; i++ {
			in := port.NewOut()
			in.Link(n.In(node.MultiPort(node.PortIn, i)))
			ins = append(ins, in)
		}

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		proc := process.New()
		defer proc.Close()

		inWriters := make([]*port.Writer, len(ins))
		for i, in := range ins {
			inWriters[i] = in.Open(proc)
		}
		outReader := out.Open(proc)

		var inPayloads []primitive.Value
		combined := []map[string]string{{}}
		for range inWriters {
			key := faker.UUIDHyphenated()
			value := faker.UUIDHyphenated()

			inPayloads = append(inPayloads, primitive.NewSlice(primitive.NewMap(primitive.NewString(key), primitive.NewString(value))))
			combined[0][key] = value
		}

		for i, inWriter := range inWriters {
			inPck := packet.New(inPayloads[i])
			inWriter.Write(inPck)
		}

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-outReader.Read():
			assert.Equal(t, combined, outPck.Payload().Interface())
			outReader.Receive(outPck)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		for _, inWriter := range inWriters {
			select {
			case backPck := <-inWriter.Receive():
				assert.NotNil(t, backPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}
		}
	})
}

func TestCombineNodeCodec_Decode(t *testing.T) {
	codec := NewCombineNodeCodec()

	spec := &CombineNodeSpec{
		Depth:   0,
		Inplace: false,
	}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func BenchmarkCombineNode_SendAndReceive(b *testing.B) {
	n := NewCombineNode()
	defer n.Close()

	var ins []*port.OutPort
	for i := 0; i < 4; i++ {
		in := port.NewOut()
		in.Link(n.In(node.MultiPort(node.PortIn, i)))
		ins = append(ins, in)
	}

	out := port.NewIn()
	n.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Close()

	inWriters := make([]*port.Writer, len(ins))
	for i, in := range ins {
		inWriters[i] = in.Open(proc)
	}
	outReader := out.Open(proc)

	var inPayloads []primitive.Value
	for range inWriters {
		inPayloads = append(inPayloads, primitive.NewString(faker.UUIDHyphenated()))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for i, inWriter := range inWriters {
			inPck := packet.New(inPayloads[i])
			inWriter.Write(inPck)
		}
		outPck := <-outReader.Read()
		outReader.Receive(outPck)
		for _, inWriter := range inWriters {
			<-inWriter.Receive()
		}
	}
}
