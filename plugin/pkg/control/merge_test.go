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

func TestNewMergeNode(t *testing.T) {
	n := NewMergeNode()
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestMergeNode_SendAndReceive(t *testing.T) {
	n := NewMergeNode()
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

	for i, inStream := range inStreams {
		inPck := packet.New(inPayloads[i])
		inStream.Send(inPck)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case outPck := <-outStream.Receive():
		assert.Equal(t, primitive.NewSlice(inPayloads...).Interface(), outPck.Payload().Interface())
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func BenchmarkMergeNode_SendAndReceive(b *testing.B) {
	n := NewMergeNode()
	defer n.Close()

	var ins []*port.Port
	for i := 0; i < 2; i++ {
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
