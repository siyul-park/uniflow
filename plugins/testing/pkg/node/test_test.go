package node

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
	"github.com/stretchr/testify/require"
)

func TestNewTestNodeCodec_Compile(t *testing.T) {
	codec := NewTestNodeCodec()
	require.NotNil(t, codec)

	spec := &TestNodeSpec{}

	n, err := codec.Compile(spec)
	require.NoError(t, err)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestNewTestNode(t *testing.T) {
	n := NewTestNode()
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestTestNode_Port(t *testing.T) {
	n := NewTestNode()
	defer n.Close()

	require.NotNil(t, n.Out(node.PortOut))
	require.NotNil(t, n.Out(node.PortWithIndex(node.PortOut, 0)))
	require.NotNil(t, n.Out(node.PortWithIndex(node.PortOut, 1)))
}

func TestTestNode_SendAndReceive(t *testing.T) {
	t.Run("SingleOutput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
		defer cancel()

		n1 := NewTestNode()
		defer n1.Close()

		var count atomic.Int32
		n2 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			count.Add(1)
			return inPck, nil
		})
		defer n2.Close()

		n1.Out(node.PortWithIndex(node.PortOut, 0)).Link(n2.In(node.PortIn))

		tester := testing2.NewTester("")
		defer tester.Exit(nil)

		go n1.Run(tester)

		select {
		case <-tester.Done():
			require.Equal(t, int32(1), count.Load())
			require.ErrorIs(t, tester.Err(), context.Canceled)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("SingleError", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
		defer cancel()

		n1 := NewTestNode()
		defer n1.Close()

		var count atomic.Int32
		n2 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			count.Add(1)
			return nil, packet.New(types.NewError(errors.New(faker.Sentence())))
		})
		defer n2.Close()

		n1.Out(node.PortWithIndex(node.PortOut, 0)).Link(n2.In(node.PortIn))

		tester := testing2.NewTester("")
		defer tester.Exit(nil)

		go n1.Run(tester)

		select {
		case <-tester.Done():
			require.Equal(t, int32(1), count.Load())
			require.Error(t, tester.Err())
			require.NotErrorIs(t, tester.Err(), context.Canceled)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("MultiOutput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
		defer cancel()

		n1 := NewTestNode()
		defer n1.Close()

		var count atomic.Int32

		n2 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			count.Add(1)
			return packet.New(types.NewString(faker.Sentence())), nil
		})
		defer n2.Close()

		n3 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			count.Add(1)

			inPayload, ok := inPck.Payload().(types.Slice)
			require.True(t, ok)
			require.Equal(t, 2, inPayload.Len())
			require.Equal(t, types.KindString, inPayload.Get(0).Kind())
			require.Equal(t, types.KindInt, inPayload.Get(1).Kind())

			return inPck, nil
		})
		defer n3.Close()

		n1.Out(node.PortWithIndex(node.PortOut, 0)).Link(n2.In(node.PortIn))
		n1.Out(node.PortWithIndex(node.PortOut, 1)).Link(n3.In(node.PortIn))

		tester := testing2.NewTester("")
		defer tester.Exit(nil)

		go n1.Run(tester)

		select {
		case <-tester.Done():
			require.Equal(t, int32(2), count.Load())
			require.ErrorIs(t, tester.Err(), context.Canceled)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("MultiError", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
		defer cancel()

		n1 := NewTestNode()
		defer n1.Close()

		var count atomic.Int32

		n2 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			count.Add(1)
			return packet.New(types.NewString(faker.Sentence())), nil
		})
		defer n2.Close()

		n3 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			count.Add(1)

			inPayload, ok := inPck.Payload().(types.Slice)
			require.True(t, ok)
			require.Equal(t, 2, inPayload.Len())
			require.Equal(t, types.KindString, inPayload.Get(0).Kind())
			require.Equal(t, types.KindInt, inPayload.Get(1).Kind())

			return nil, packet.New(types.NewError(errors.New(faker.Sentence())))
		})
		defer n3.Close()

		n1.Out(node.PortWithIndex(node.PortOut, 0)).Link(n2.In(node.PortIn))
		n1.Out(node.PortWithIndex(node.PortOut, 1)).Link(n3.In(node.PortIn))

		tester := testing2.NewTester("")
		defer tester.Exit(nil)

		go n1.Run(tester)

		select {
		case <-tester.Done():
			require.Equal(t, int32(2), count.Load())
			require.Error(t, tester.Err())
			require.NotErrorIs(t, tester.Err(), context.Canceled)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})
}
