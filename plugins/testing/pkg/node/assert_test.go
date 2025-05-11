package node

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/meta"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
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

func TestAssertNode_SendAndReceive(t *testing.T) {
	t.Run("DirectAssert", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()

		agent := runtime.NewAgent()
		defer agent.Close()

		n1 := NewTestNode()
		defer n1.Close()

		n2 := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      "snippet",
				Namespace: meta.DefaultNamespace,
				Name:      "target",
			},
			Node: node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
				return packet.New(types.NewInt(10)), nil
			}),
		}
		defer n2.Close()

		evaluator := func(_ context.Context, payload interface{}) (bool, error) {
			val, ok := payload.(types.Int)
			if !ok {
				return false, nil
			}
			return val.Int() == 10, nil
		}

		n3 := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      "assert",
				Namespace: meta.DefaultNamespace,
				Name:      "assert",
			},
			Node: NewAssertNode(&AssertNodeSpec{
				Expect: "self == 10",
			}, agent, evaluator),
		}
		defer n3.Close()

		agent.Load(n2)
		defer agent.Unload(n2)

		agent.Load(n3)
		defer agent.Unload(n3)

		n1.Out(node.PortWithIndex(node.PortOut, 0)).Link(n2.In(node.PortIn))
		n1.Out(node.PortWithIndex(node.PortOut, 1)).Link(n3.In(node.PortIn))

		tester := testing2.NewTester("")
		go n1.Run(tester)

		select {
		case <-tester.Done():
			require.ErrorIs(t, tester.Err(), context.Canceled)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("AssertFailed", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()

		agent := runtime.NewAgent()
		defer agent.Close()

		n1 := NewTestNode()
		defer n1.Close()

		n2 := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      "snippet",
				Namespace: meta.DefaultNamespace,
				Name:      "target",
			},
			Node: node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
				return packet.New(types.NewInt(5)), nil
			}),
		}
		defer n2.Close()

		evaluator := func(_ context.Context, payload interface{}) (bool, error) {
			val, ok := payload.(types.Int)
			if !ok {
				return false, nil
			}
			return val.Int() == 10, nil
		}

		n3 := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      "assert",
				Namespace: meta.DefaultNamespace,
				Name:      "assert",
			},
			Node: NewAssertNode(&AssertNodeSpec{
				Expect: "self == 10",
			}, agent, evaluator),
		}
		defer n3.Close()

		agent.Load(n2)
		defer agent.Unload(n2)

		agent.Load(n3)
		defer agent.Unload(n3)

		n1.Out(node.PortWithIndex(node.PortOut, 0)).Link(n2.In(node.PortIn))
		n1.Out(node.PortWithIndex(node.PortOut, 1)).Link(n3.In(node.PortIn))

		tester := testing2.NewTester("")
		go n1.Run(tester)

		select {
		case <-tester.Done():
			require.Error(t, tester.Err())
			require.NotErrorIs(t, tester.Err(), context.Canceled)
			require.Contains(t, tester.Err().Error(), "assertion failed")
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("TargetAssert", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()

		agent := runtime.NewAgent()
		defer agent.Close()

		n1 := NewTestNode()
		defer n1.Close()

		targetNodeName := "target"

		n2 := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      "snippet",
				Namespace: meta.DefaultNamespace,
				Name:      targetNodeName,
			},
			Node: node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
				return packet.New(types.NewInt(10)), nil
			}),
		}
		defer n2.Close()

		n3 := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      "snippet",
				Namespace: meta.DefaultNamespace,
				Name:      "non-target",
			},
			Node: node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
				return packet.New(types.NewInt(20)), nil
			}),
		}
		defer n3.Close()

		evaluator := func(_ context.Context, payload interface{}) (bool, error) {
			val, ok := payload.(types.Int)
			if !ok {
				return false, nil
			}
			return val.Int() == 10, nil
		}

		n4 := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      "assert",
				Namespace: meta.DefaultNamespace,
				Name:      "assert",
			},
			Node: NewAssertNode(&AssertNodeSpec{
				Expect: "self == 10",
				Target: &AssertNodeTarget{
					Name: targetNodeName,
					Port: node.PortOut,
				},
			}, agent, evaluator),
		}
		defer n4.Close()

		agent.Load(n2)
		defer agent.Unload(n2)

		agent.Load(n3)
		defer agent.Unload(n3)

		agent.Load(n4)
		defer agent.Unload(n4)

		n1.Out(node.PortWithIndex(node.PortOut, 0)).Link(n2.In(node.PortIn))
		n1.Out(node.PortWithIndex(node.PortOut, 1)).Link(n4.In(node.PortIn))
		n2.Out(node.PortOut).Link(n3.In(node.PortIn))

		tester := testing2.NewTester("")
		go n1.Run(tester)

		select {
		case <-tester.Done():
			require.ErrorIs(t, tester.Err(), context.Canceled)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("TargetNotFound", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()

		agent := runtime.NewAgent()
		defer agent.Close()

		n1 := NewTestNode()
		defer n1.Close()

		n2 := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      "snippet",
				Namespace: meta.DefaultNamespace,
				Name:      "target",
			},
			Node: node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
				return packet.New(types.NewInt(10)), nil
			}),
		}
		defer n2.Close()

		evaluator := func(_ context.Context, payload interface{}) (bool, error) {
			return true, nil
		}

		n3 := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      "assert",
				Namespace: meta.DefaultNamespace,
				Name:      "assert",
			},
			Node: NewAssertNode(&AssertNodeSpec{
				Expect: "self == 10",
				Target: &AssertNodeTarget{
					Name: "non-existent-node",
					Port: node.PortOut,
				},
			}, agent, evaluator),
		}
		defer n3.Close()

		agent.Load(n2)
		defer agent.Unload(n2)

		agent.Load(n3)
		defer agent.Unload(n3)

		n1.Out(node.PortWithIndex(node.PortOut, 0)).Link(n2.In(node.PortIn))
		n1.Out(node.PortWithIndex(node.PortOut, 1)).Link(n3.In(node.PortIn))

		tester := testing2.NewTester("")
		go n1.Run(tester)

		select {
		case <-tester.Done():
			require.Error(t, tester.Err())
			require.NotErrorIs(t, tester.Err(), context.Canceled)
			require.Contains(t, tester.Err().Error(), "target symbol not found")
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})
}
