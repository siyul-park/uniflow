package node

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/runtime"
	testing2 "github.com/siyul-park/uniflow/pkg/testing"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestNewAssertNodeCodec(t *testing.T) {
	codec := NewAssertNodeCodec(nil, nil)
	require.NotNil(t, codec)
}

func TestAssertNode_Port(t *testing.T) {
	n := NewAssertNode(nil, nil, nil)
	defer n.Close()

	require.NotNil(t, n.In(node.PortIn))
	require.NotNil(t, n.Out(node.PortOut))
}

func TestAssertNode_Evaluate(t *testing.T) {
	t.Run("DirectAssert", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()

		agent := runtime.NewAgent()
		defer agent.Close()

		n1 := NewTestNode()
		defer n1.Close()

		n2 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return packet.New(types.NewInt(10)), nil
		})
		defer n2.Close()

		n3 := NewAssertNode(nil, agent, nil)
		defer n3.Close()

		n1.Out(node.PortWithIndex(node.PortOut, 0)).Link(n2.In(node.PortIn))
		n1.Out(node.PortWithIndex(node.PortOut, 1)).Link(n3.In(node.PortIn))

		tester := testing2.NewTester("")

		go n1.Run(tester)

		select {
		case <-tester.Done():
			require.NoError(t, tester.Err())
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("AssertFailed", func(t *testing.T) {

	})

	t.Run("TargetAssert", func(t *testing.T) {

	})

	t.Run("TargetNotFound", func(t *testing.T) {

	})
}
