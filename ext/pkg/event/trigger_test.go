package event

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestTriggerNodeCodec_Decode(t *testing.T) {
	broker := NewBroker()

	topic := faker.UUIDHyphenated()

	codec := NewTriggerNodeCodec(broker, broker)

	spec := &TriggerNodeSpec{
		Topic: topic,
	}

	n, err := codec.Compile(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestNewTriggerNode(t *testing.T) {
	q := NewQueue(0)
	c := NewConsumer(q)
	p := NewProducer(q)

	n := NewTriggerNode(p, c)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestTriggerNode_Port(t *testing.T) {
	q := NewQueue(0)
	c := NewConsumer(q)
	p := NewProducer(q)

	n := NewTriggerNode(p, c)
	defer n.Close()

	assert.NotNil(t, n.In(node.PortIn))
	assert.NotNil(t, n.Out(node.PortOut))
	assert.NotNil(t, n.Out(node.PortErr))
}

func TestTriggerNode_SendAndReceive(t *testing.T) {
	t.Run("NoInputToSingleOutput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
		defer cancel()

		q := NewQueue(0)
		c := NewConsumer(q)
		p := NewProducer(q)

		n := NewTriggerNode(p, c)
		defer n.Close()

		n.Listen()

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		count := 0

		out.Accept(port.ListenFunc(func(proc *process.Process) {
			outReader := out.Open(proc)

			for {
				_, ok := <-outReader.Read()
				if !ok {
					return
				}
				count += 1
				outReader.Receive(packet.None)
			}
		}))

		e := New(nil)
		q.Push(e)

		select {
		case <-e.Done():
			assert.Equal(t, 1, count)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("SingleInputToNoOutput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		q := NewQueue(0)
		c := NewConsumer(q)
		p := NewProducer(q)

		n := NewTriggerNode(p, c)
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)

		inPayload := types.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case <-inWriter.Receive():
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		select {
		case e := <-c.Consume():
			assert.Equal(t, inPayload.Interface(), e.Data())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})
}

func BenchmarkTriggerNode_SendAndReceive(b *testing.B) {
	q := NewQueue(0)
	c := NewConsumer(q)
	p := NewProducer(q)

	n := NewTriggerNode(p, c)
	defer n.Close()

	n.Listen()

	out := port.NewIn()
	n.Out(node.PortOut).Link(out)

	out.Accept(port.ListenFunc(func(proc *process.Process) {
		outReader := out.Open(proc)

		for {
			_, ok := <-outReader.Read()
			if !ok {
				return
			}
			outReader.Receive(packet.None)
		}
	}))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e := New(nil)
		q.Push(e)
		<-e.Done()
	}
}
