package system

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/event"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewTriggerNode(t *testing.T) {
	q := event.NewQueue(0)
	c := event.NewConsumer(q)

	n := NewTriggerNode(c)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestTriggerNode_Port(t *testing.T) {
	q := event.NewQueue(0)
	c := event.NewConsumer(q)

	n := NewTriggerNode(c)
	defer n.Close()

	assert.NotNil(t, n.In(node.PortIn))
	assert.NotNil(t, n.Out(node.PortOut))
	assert.NotNil(t, n.Out(node.PortErr))
}

func TestTriggerNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	q := event.NewQueue(0)
	c := event.NewConsumer(q)

	n := NewTriggerNode(c)
	defer n.Close()

	out := port.NewIn()
	n.Out(node.PortOut).Link(out)

	count := 0

	out.AddHandler(port.HandlerFunc(func(proc *process.Process) {
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
	defer e.Close()

	q.Push(e)

	n.Listen()

	select {
	case <-e.Done():
		assert.Equal(t, 1, count)
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}
