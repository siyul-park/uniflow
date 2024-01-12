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
			for _, inPck := range inPcks {
				if inPck == nil {
					return nil, nil
				}
			}
			return inPcks[len(inPcks)-1], nil
		})
		defer n.Close()

		var ins []*port.Port
		for i := 0; i < 4; i++ {
			in := port.New()
			inPort, _ := n.Port(MultiPort(PortIn, i))
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

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-outStream.Receive():
			assert.Equal(t, inPayload, outPck.Payload())
			outStream.Send(outPck)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		for _, inStream := range inStreams {
			select {
			case backPck := <-inStream.Receive():
				assert.NotNil(t, backPck)
			case <-ctx.Done():
				assert.Fail(t, ctx.Err().Error())
			}
		}
	})

	t.Run("With Err Port", func(t *testing.T) {
		n := NewManyToOneNode(func(_ *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
			for _, inPck := range inPcks {
				if inPck == nil {
					return nil, nil
				}
			}
			return nil, inPcks[len(inPcks)-1]
		})
		defer n.Close()

		var ins []*port.Port
		for i := 0; i < 4; i++ {
			in := port.New()
			inPort, _ := n.Port(MultiPort(PortIn, i))
			inPort.Link(in)

			ins = append(ins, in)
		}

		err := port.New()
		errPort, _ := n.Port(PortErr)
		errPort.Link(err)

		proc := process.New()
		defer proc.Exit(nil)

		var inStreams []*port.Stream
		for _, in := range ins {
			inStreams = append(inStreams, in.Open(proc))
		}
		errStream := err.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())

		for _, inStream := range inStreams {
			inPck := packet.New(inPayload)
			inStream.Send(inPck)
		}

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case errPck := <-errStream.Receive():
			assert.Equal(t, inPayload, errPck.Payload())
			errStream.Send(errPck)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		for _, inStream := range inStreams {
			select {
			case backPck := <-inStream.Receive():
				assert.NotNil(t, backPck)
			case <-ctx.Done():
				assert.Fail(t, ctx.Err().Error())
			}
		}
	})
}

func BenchmarkManyToOneNode_SendAndReceive(b *testing.B) {
	n := NewManyToOneNode(func(_ *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
		for _, inPck := range inPcks {
			if inPck == nil {
				return nil, nil
			}
		}
		return inPcks[len(inPcks)-1], nil
	})
	defer n.Close()

	var ins []*port.Port
	for i := 0; i < 2; i++ {
		in := port.New()
		inPort, _ := n.Port(MultiPort(PortIn, i))
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

	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			for _, inStream := range inStreams {
				inPck := packet.New(inPayload)
				inStream.Send(inPck)
			}
			<-outStream.Receive()
		}
	})
}
