package control

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/event"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewTriggerNode(t *testing.T) {
	q := event.NewQueue(0)
	c := event.NewConsumer(q)
	p := event.NewProducer(q)

	n := NewTriggerNode(p, c)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestTriggerNode_Port(t *testing.T) {
	q := event.NewQueue(0)
	c := event.NewConsumer(q)
	p := event.NewProducer(q)

	n := NewTriggerNode(p, c)
	defer n.Close()

	assert.NotNil(t, n.In(node.PortIn))
	assert.NotNil(t, n.Out(node.PortOut))
	assert.NotNil(t, n.Out(node.PortErr))
}

func TestTriggerNode_SendAndReceive(t *testing.T) {
	t.Run("Out", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		q := event.NewQueue(0)
		c := event.NewConsumer(q)
		p := event.NewProducer(q)

		n := NewTriggerNode(p, c)
		defer n.Close()

		n.Listen()

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		count := 0

		out.AddInitHook(port.InitHookFunc(func(proc *process.Process) {
			outReader := out.Open(proc)

			for {
				outPck, ok := <-outReader.Read()
				if !ok {
					return
				}
				count += 1
				proc.Stack().Clear(outPck)
			}
		}))

		e := event.New(nil)
		q.Push(e)

		select {
		case <-e.Done():
			assert.Equal(t, 1, count)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("In", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		q := event.NewQueue(0)
		c := event.NewConsumer(q)
		p := event.NewProducer(q)

		n := NewTriggerNode(p, c)
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		proc := process.New()
		defer proc.Close()

		inWriter := in.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case <-proc.Stack().Done(inPck):
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

func TestTriggerNodeCodec_Decode(t *testing.T) {
	broker := event.NewBroker()

	topic := faker.UUIDHyphenated()

	codec := NewTriggerNodeCodec(broker)

	spec := &TriggerNodeSpec{
		Topic: topic,
	}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func BenchmarkTriggerNode_SendAndReceive(b *testing.B) {
	q := event.NewQueue(0)
	c := event.NewConsumer(q)
	p := event.NewProducer(q)

	n := NewTriggerNode(p, c)
	defer n.Close()

	n.Listen()

	out := port.NewIn()
	n.Out(node.PortOut).Link(out)

	out.AddInitHook(port.InitHookFunc(func(proc *process.Process) {
		outReader := out.Open(proc)

		for {
			outPck, ok := <-outReader.Read()
			if !ok {
				return
			}
			proc.Stack().Clear(outPck)
		}
	}))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e := event.New(nil)
		q.Push(e)
		<-e.Done()
	}
}
