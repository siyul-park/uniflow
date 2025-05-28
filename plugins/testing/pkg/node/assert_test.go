package node

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/language/text"
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
	compiler := text.NewCompiler()
	agent := runtime.NewAgent()
	defer agent.Close()

	codec := NewAssertNodeCodec(compiler, agent)
	require.NotNil(t, codec)
}

func TestAssertNode_Port(t *testing.T) {
	n := NewAssertNode(nil)
	defer n.Close()

	require.NotNil(t, n.In(node.PortIn))
	require.NotNil(t, n.Out(node.PortOut))
}

func TestAssertNode_SendAndReceive(t *testing.T) {
	t.Run("DirectAssert", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		expect := func(_ context.Context, payload any) (bool, error) {
			return payload == 10, nil
		}

		n1 := NewTestNode()
		defer n1.Close()

		n2 := node.NewOneToOneNode(func(_ *process.Process, _ *packet.Packet) (*packet.Packet, *packet.Packet) {
			return packet.New(types.NewInt(10)), nil
		})
		defer n2.Close()

		n3 := NewAssertNode(expect)
		defer n3.Close()

		n1.Out(node.PortWithIndex(node.PortOut, 0)).Link(n2.In(node.PortIn))
		n1.Out(node.PortWithIndex(node.PortOut, 1)).Link(n3.In(node.PortIn))

		tester := testing2.NewTester("")
		defer tester.Exit(nil)

		go n1.Run(tester)

		select {
		case <-tester.Done():
			require.ErrorIs(t, tester.Err(), context.Canceled)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("AssertFailed", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		expect := func(_ context.Context, payload any) (bool, error) {
			return payload == 10, nil
		}

		n1 := NewTestNode()
		defer n1.Close()

		n2 := node.NewOneToOneNode(func(_ *process.Process, _ *packet.Packet) (*packet.Packet, *packet.Packet) {
			return packet.New(types.NewInt(20)), nil
		})
		defer n2.Close()

		n3 := NewAssertNode(expect)
		defer n3.Close()

		n1.Out(node.PortWithIndex(node.PortOut, 0)).Link(n2.In(node.PortIn))
		n1.Out(node.PortWithIndex(node.PortOut, 1)).Link(n3.In(node.PortIn))

		tester := testing2.NewTester("")
		defer tester.Exit(nil)

		go n1.Run(tester)

		select {
		case <-tester.Done():
			require.ErrorIs(t, tester.Err(), ErrAssertFail)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("IDAssert", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		agent := runtime.NewAgent()
		defer agent.Close()

		expect := func(_ context.Context, payload any) (bool, error) {
			return payload == 10, nil
		}

		tester := testing2.NewTester("")
		defer tester.Exit(nil)

		n1 := NewTestNode()
		defer n1.Close()

		target := uuid.Must(uuid.NewV7())
		n2 := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        target,
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(func(_ *process.Process, _ *packet.Packet) (*packet.Packet, *packet.Packet) {
				return packet.New(types.NewInt(10)), nil
			}),
		}
		defer n2.Close()

		n3 := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(func(_ *process.Process, _ *packet.Packet) (*packet.Packet, *packet.Packet) {
				return packet.New(types.NewInt(20)), nil
			}),
		}
		defer n3.Close()

		n4 := NewAssertNode(expect)
		defer n4.Close()

		n4.SetTarget(func(proc *process.Process, payload any, index int) (any, int, error) {
			if index < 0 {
				index = 0
			}

			frames := agent.Frames(proc.ID())
			for i := index; i < len(frames); i++ {
				frame := frames[i]
				sym := frame.Symbol
				if sym.ID() != target {
					continue
				}

				if frame.OutPort != nil && frame.OutPort == sym.Out(node.PortOut) {
					if frame.OutPck == nil || frame.OutPck.Payload() == nil {
						continue
					}
					return types.InterfaceOf(frame.OutPck.Payload()), i, nil
				}
			}

			return nil, 0, errors.WithStack(ErrAssertFail)
		})

		n2.In(node.PortIn)
		n2.Out(node.PortOut)
		n3.In(node.PortIn)
		n3.Out(node.PortOut)

		agent.Load(n2)
		defer agent.Unload(n2)

		agent.Load(n3)
		defer agent.Unload(n3)

		n1.Out(node.PortWithIndex(node.PortOut, 0)).Link(n2.In(node.PortIn))
		n1.Out(node.PortWithIndex(node.PortOut, 1)).Link(n4.In(node.PortIn))
		n2.Out(node.PortOut).Link(n3.In(node.PortIn))

		go n1.Run(tester)

		select {
		case <-tester.Done():
			require.ErrorIs(t, tester.Err(), context.Canceled)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("NameAssert", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		agent := runtime.NewAgent()
		defer agent.Close()

		expect := func(_ context.Context, payload any) (bool, error) {
			return payload == 10, nil
		}

		tester := testing2.NewTester("")
		defer tester.Exit(nil)

		n1 := NewTestNode()
		defer n1.Close()

		target := faker.UUIDHyphenated()
		n2 := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      target,
			},
			Node: node.NewOneToOneNode(func(_ *process.Process, _ *packet.Packet) (*packet.Packet, *packet.Packet) {
				return packet.New(types.NewInt(10)), nil
			}),
		}
		defer n2.Close()

		nonTarget := uuid.Must(uuid.NewV7())
		n3 := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        nonTarget,
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(func(_ *process.Process, _ *packet.Packet) (*packet.Packet, *packet.Packet) {
				return packet.New(types.NewInt(20)), nil
			}),
		}
		defer n3.Close()

		n4 := NewAssertNode(expect)
		defer n4.Close()

		n4.SetTarget(func(proc *process.Process, payload any, index int) (any, int, error) {
			if index < 0 {
				index = 0
			}

			frames := agent.Frames(proc.ID())
			for i := index; i < len(frames); i++ {
				frame := frames[i]
				sym := frame.Symbol
				if sym.Name() != target {
					continue
				}

				if frame.OutPort != nil && frame.OutPort == sym.Out(node.PortOut) {
					if frame.OutPck == nil || frame.OutPck.Payload() == nil {
						continue
					}
					return types.InterfaceOf(frame.OutPck.Payload()), i, nil
				}
			}

			return nil, 0, errors.WithStack(ErrAssertFail)
		})

		n2.In(node.PortIn)
		n2.Out(node.PortOut)
		n3.In(node.PortIn)
		n3.Out(node.PortOut)

		agent.Load(n2)
		defer agent.Unload(n2)

		agent.Load(n3)
		defer agent.Unload(n3)

		n1.Out(node.PortWithIndex(node.PortOut, 0)).Link(n2.In(node.PortIn))
		n1.Out(node.PortWithIndex(node.PortOut, 1)).Link(n4.In(node.PortIn))
		n2.Out(node.PortOut).Link(n3.In(node.PortIn))

		go n1.Run(tester)

		select {
		case <-tester.Done():
			require.ErrorIs(t, tester.Err(), context.Canceled)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("IDNotFound", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		agent := runtime.NewAgent()
		defer agent.Close()

		expect := func(_ context.Context, payload any) (bool, error) {
			return payload == 10, nil
		}

		tester := testing2.NewTester("")
		defer tester.Exit(nil)

		n1 := NewTestNode()
		defer n1.Close()

		n2 := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(func(_ *process.Process, _ *packet.Packet) (*packet.Packet, *packet.Packet) {
				return packet.New(types.NewInt(10)), nil
			}),
		}
		defer n2.Close()

		nonTarget := uuid.Must(uuid.NewV7())
		n3 := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        nonTarget,
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(func(_ *process.Process, _ *packet.Packet) (*packet.Packet, *packet.Packet) {
				return packet.New(types.NewInt(20)), nil
			}),
		}
		defer n3.Close()

		n4 := NewAssertNode(expect)
		defer n4.Close()

		n4.SetTarget(func(proc *process.Process, payload any, index int) (any, int, error) {
			if index < 0 {
				index = 0
			}

			target := uuid.Must(uuid.NewV7())
			frames := agent.Frames(proc.ID())
			for i := index; i < len(frames); i++ {
				frame := frames[i]
				sym := frame.Symbol
				if sym.ID() != target {
					continue
				}

				if frame.OutPort != nil && frame.OutPort == sym.Out(node.PortOut) {
					if frame.OutPck == nil || frame.OutPck.Payload() == nil {
						continue
					}
					return types.InterfaceOf(frame.OutPck.Payload()), i, nil
				}
			}

			return nil, 0, errors.WithStack(ErrAssertFail)
		})

		n2.In(node.PortIn)
		n2.Out(node.PortOut)
		n3.In(node.PortIn)
		n3.Out(node.PortOut)

		agent.Load(n2)
		defer agent.Unload(n2)

		agent.Load(n3)
		defer agent.Unload(n3)

		n1.Out(node.PortWithIndex(node.PortOut, 0)).Link(n2.In(node.PortIn))
		n1.Out(node.PortWithIndex(node.PortOut, 1)).Link(n4.In(node.PortIn))
		n2.Out(node.PortOut).Link(n3.In(node.PortIn))

		go n1.Run(tester)

		select {
		case <-tester.Done():
			require.ErrorIs(t, tester.Err(), ErrAssertFail)
			require.NotErrorIs(t, tester.Err(), context.Canceled)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("NamespaceDiff", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		agent := runtime.NewAgent()
		defer agent.Close()

		expect := func(_ context.Context, payload any) (bool, error) {
			return payload == 10, nil
		}

		tester := testing2.NewTester("")
		defer tester.Exit(nil)

		n1 := NewTestNode()
		defer n1.Close()

		target := uuid.Must(uuid.NewV7())
		n2 := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        target,
				Kind:      faker.UUIDHyphenated(),
				Namespace: "TargetNamespace",
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(func(_ *process.Process, _ *packet.Packet) (*packet.Packet, *packet.Packet) {
				return packet.New(types.NewInt(10)), nil
			}),
		}
		defer n2.Close()

		nonTarget := uuid.Must(uuid.NewV7())
		n3 := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        nonTarget,
				Kind:      faker.UUIDHyphenated(),
				Namespace: "NonTargetNamespace",
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(func(_ *process.Process, _ *packet.Packet) (*packet.Packet, *packet.Packet) {
				return packet.New(types.NewInt(20)), nil
			}),
		}
		defer n3.Close()

		n4 := NewAssertNode(expect)
		defer n4.Close()

		n4.SetTarget(func(proc *process.Process, payload any, index int) (any, int, error) {
			if index < 0 {
				index = 0
			}

			frames := agent.Frames(proc.ID())
			for i := index; i < len(frames); i++ {
				frame := frames[i]
				sym := frame.Symbol
				if sym.Namespace() != meta.DefaultNamespace || sym.ID() != target {
					continue
				}

				if frame.OutPort != nil && frame.OutPort == sym.Out(node.PortOut) {
					if frame.OutPck == nil || frame.OutPck.Payload() == nil {
						continue
					}
					return types.InterfaceOf(frame.OutPck.Payload()), i, nil
				}
			}

			return nil, 0, errors.WithStack(ErrAssertFail)
		})

		n2.In(node.PortIn)
		n2.Out(node.PortOut)
		n3.In(node.PortIn)
		n3.Out(node.PortOut)

		agent.Load(n2)
		defer agent.Unload(n2)

		agent.Load(n3)
		defer agent.Unload(n3)

		n1.Out(node.PortWithIndex(node.PortOut, 0)).Link(n2.In(node.PortIn))
		n1.Out(node.PortWithIndex(node.PortOut, 1)).Link(n4.In(node.PortIn))
		n2.Out(node.PortOut).Link(n3.In(node.PortIn))

		go n1.Run(tester)

		select {
		case <-tester.Done():
			require.ErrorIs(t, tester.Err(), ErrAssertFail)
			require.NotErrorIs(t, tester.Err(), context.Canceled)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})
}
