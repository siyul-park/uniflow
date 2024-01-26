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

func TestCombineNodeCodec_Decode(t *testing.T) {
	codec := NewCombineNodeCodec()

	spec := &CombineNodeSpec{
		Depth: 0,
	}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
}

func TestNewCombineNode(t *testing.T) {
	n := NewCombineNode(0)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestCombineNode_SendAndReceive(t *testing.T) {
	t.Run("depth = 0", func(t *testing.T) {
		n := NewCombineNode(0)
		defer n.Close()

		var ins []*port.Port
		for i := 0; i < 4; i++ {
			in := port.New()
			inPort, _ := n.Port(node.MultiPort(node.PortIn, i))
			inPort.Link(in)

			ins = append(ins, in)
		}

		out := port.New()
		outPort, _ := n.Port(node.PortOut)
		outPort.Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		var inStreams []*port.Stream
		for _, in := range ins {
			inStreams = append(inStreams, in.Open(proc))
		}
		outStream := out.Open(proc)

		var inPayloads []primitive.Value
		for range inStreams {
			inPayloads = append(inPayloads, primitive.NewString(faker.UUIDHyphenated()))
		}

		combined := primitive.NewSlice(inPayloads...).Interface()

		for i, inStream := range inStreams {
			inPck := packet.New(inPayloads[i])
			inStream.Send(inPck)
		}

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-outStream.Receive():
			assert.Equal(t, combined, outPck.Payload().Interface())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("depth = 1", func(t *testing.T) {
		n := NewCombineNode(1)
		defer n.Close()

		var ins []*port.Port
		for i := 0; i < 4; i++ {
			in := port.New()
			inPort, _ := n.Port(node.MultiPort(node.PortIn, i))
			inPort.Link(in)

			ins = append(ins, in)
		}

		out := port.New()
		outPort, _ := n.Port(node.PortOut)
		outPort.Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		var inStreams []*port.Stream
		for _, in := range ins {
			inStreams = append(inStreams, in.Open(proc))
		}
		outStream := out.Open(proc)

		var inPayloads []primitive.Value
		for range inStreams {
			inPayloads = append(inPayloads, primitive.NewString(faker.UUIDHyphenated()))
		}

		combined := inPayloads[len(inPayloads)-1].Interface()

		for i, inStream := range inStreams {
			inPck := packet.New(inPayloads[i])
			inStream.Send(inPck)
		}

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-outStream.Receive():
			assert.Equal(t, combined, outPck.Payload().Interface())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})
}

func BenchmarkCombineNode_SendAndReceive(b *testing.B) {
	n := NewCombineNode(0)
	defer n.Close()

	var ins []*port.Port
	for i := 0; i < 4; i++ {
		in := port.New()
		inPort, _ := n.Port(node.MultiPort(node.PortIn, i))
		inPort.Link(in)

		ins = append(ins, in)
	}

	out := port.New()
	outPort, _ := n.Port(node.PortOut)
	outPort.Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	var inStreams []*port.Stream
	for _, in := range ins {
		inStreams = append(inStreams, in.Open(proc))
	}
	outStream := out.Open(proc)

	var inPayloads []primitive.Value
	for range inStreams {
		inPayloads = append(inPayloads, primitive.NewString(faker.UUIDHyphenated()))
	}

	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			for i, inStream := range inStreams {
				inPck := packet.New(inPayloads[i])
				inStream.Send(inPck)
			}
			<-outStream.Receive()
		}
	})
}
