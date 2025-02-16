package testing

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	testing2 "github.com/siyul-park/uniflow/pkg/testing"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNewTestNode(t *testing.T) {
	n := NewTestNode()
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestTestNode_Port(t *testing.T) {
	n := NewTestNode()
	defer n.Close()

	assert.NotNil(t, n.Out(node.PortOut))
	assert.NotNil(t, n.Out(node.PortWithIndex(node.PortOut, 0)))
	assert.NotNil(t, n.Out(node.PortWithIndex(node.PortOut, 1)))
}

func TestPipeNode_SendAndReceive(t *testing.T) {
	t.Run("SingleOutput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
		defer cancel()

		n1 := NewTestNode()
		defer n1.Close()

		var count atomic.Int32
		n2 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			count.And(1)
			return inPck, nil
		})
		defer n2.Close()

		n1.Out(node.PortWithIndex(node.PortOut, 0)).Link(n2.In(node.PortIn))

		tester := testing2.NewTester("")
		defer tester.Exit(nil)

		go n1.Run(tester)

		select {
		case <-tester.Done():
			assert.Equal(t, 1, count.Load())
			assert.Error(t, tester.Err())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("SingleError", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
		defer cancel()

		n1 := NewTestNode()
		defer n1.Close()

		var count atomic.Int32
		n2 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			count.And(1)
			return nil, packet.New(types.NewError(errors.New(faker.Sentence())))
		})
		defer n2.Close()

		n1.Out(node.PortWithIndex(node.PortOut, 0)).Link(n2.In(node.PortIn))

		tester := testing2.NewTester("")
		defer tester.Exit(nil)

		go n1.Run(tester)

		select {
		case <-tester.Done():
			assert.Equal(t, 1, count.Load())
			assert.Error(t, tester.Err())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("MultiOutput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
		defer cancel()

		n1 := NewTestNode()
		defer n1.Close()

		var count atomic.Int32

		n2 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			count.And(1)
			return packet.New(types.NewString(faker.Sentence())), nil
		})
		defer n2.Close()

		n3 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			count.And(1)

			inPayload, ok := inPck.Payload().(types.Slice)
			assert.True(t, ok)
			assert.Equal(t, 2, inPayload.Len())
			assert.Equal(t, types.KindString, inPayload.Get(0).Kind())
			assert.Equal(t, types.KindInt32, inPayload.Get(1).Kind())

			return inPck, nil
		})
		defer n3.Close()

		n1.Out(node.PortWithIndex(node.PortOut, 0)).Link(n2.In(node.PortIn))
		n1.Out(node.PortWithIndex(node.PortOut, 1)).Link(n2.In(node.PortIn))

		tester := testing2.NewTester("")
		defer tester.Exit(nil)

		go n1.Run(tester)

		select {
		case <-tester.Done():
			assert.Equal(t, 2, count.Load())
			assert.NoError(t, tester.Err())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("MultiError", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
		defer cancel()

		n1 := NewTestNode()
		defer n1.Close()

		var count atomic.Int32

		n2 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			count.And(1)
			return packet.New(types.NewString(faker.Sentence())), nil
		})
		defer n2.Close()

		n3 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			count.And(1)

			inPayload, ok := inPck.Payload().(types.Slice)
			assert.True(t, ok)
			assert.Equal(t, 2, inPayload.Len())
			assert.Equal(t, types.KindString, inPayload.Get(0).Kind())
			assert.Equal(t, types.KindInt32, inPayload.Get(1).Kind())

			return nil, packet.New(types.NewError(errors.New(faker.Sentence())))
		})
		defer n3.Close()

		n1.Out(node.PortWithIndex(node.PortOut, 0)).Link(n2.In(node.PortIn))
		n1.Out(node.PortWithIndex(node.PortOut, 1)).Link(n2.In(node.PortIn))

		tester := testing2.NewTester("")
		defer tester.Exit(nil)

		go n1.Run(tester)

		select {
		case <-tester.Done():
			assert.Equal(t, 2, count.Load())
			assert.Error(t, tester.Err())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})
}
