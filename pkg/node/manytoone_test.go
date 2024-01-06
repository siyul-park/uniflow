package node

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewManyToOneNode(t *testing.T) {
	n := NewManyToOneNode(nil)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestManyToOneNode_Port(t *testing.T) {
	n := NewManyToOneNode(nil)
	defer n.Close()

	p, ok := n.Port(MultiPort(PortIn, 0))
	assert.True(t, ok)
	assert.NotNil(t, p)

	p, ok = n.Port(PortOut)
	assert.True(t, ok)
	assert.NotNil(t, p)

	p, ok = n.Port(PortErr)
	assert.True(t, ok)
	assert.NotNil(t, p)
}

func TestManyToOneNode_SendAndReceive(t *testing.T) {
	t.Run("With Out Port", func(t *testing.T) {
		n := NewManyToOneNode(func(_ *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
			if inPcks[len(inPcks)-1] != nil {
				return inPcks[len(inPcks)-1], nil
			}
			return nil, nil
		})
		defer n.Close()

		var ins []*port.Port
		for i := 0; i < 16; i++ {
			in := port.New()
			inPort, _ := n.Port(MultiPort(PortIn, 0))
			inPort.Link(in)

			ins = append(ins, in)
		}

		out := port.New()
		outPort, _ := n.Port(PortOut)
		outPort.Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		var inStreams []*port.Stream
		for _, in := range ins {
			inStreams = append(inStreams, in.Open(proc))
		}
		outStream := out.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())

		for _, inStream := range inStreams {
			inPck := packet.New(inPayload)
			inStream.Send(inPck)
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		select {
		case outPck := <-outStream.Receive():
			assert.Equal(t, inPayload, outPck.Payload())

			outStream.Send(outPck)
			for _, inStream := range inStreams {
				select {
				case outPck := <-inStream.Receive():
					assert.NotNil(t, outPck)
				case <-ctx.Done():
					assert.Fail(t, ctx.Err().Error())
				}
			}
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})
}
