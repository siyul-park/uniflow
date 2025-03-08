package system

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/require"
)

func TestSignalNodeCodec_Compile(t *testing.T) {
	opcode := faker.UUIDHyphenated()

	codec := NewSignalNodeCodec(map[string]func(context.Context) (<-chan any, error){
		opcode: func(_ context.Context) (<-chan any, error) {
			return make(chan any), nil
		},
	})

	spec := &SignalNodeSpec{
		Topic: opcode,
	}

	n, err := codec.Compile(spec)
	require.NoError(t, err)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestNewSignalNode(t *testing.T) {
	n := NewSignalNode(nil)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestSignalNode_Port(t *testing.T) {
	n := NewSignalNode(nil)
	defer n.Close()

	require.NotNil(t, n.Out(node.PortOut))
}

func TestSignalNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	signal := make(chan any)
	done := make(chan struct{})

	n := NewSignalNode(signal)
	defer n.Close()

	out := port.NewIn()
	n.Out(node.PortOut).Link(out)

	n.Listen()
	defer n.Shutdown()

	out.AddListener(port.ListenFunc(func(proc *process.Process) {
		defer close(done)

		outReader := out.Open(proc)

		select {
		case outPck := <-outReader.Read():
			outReader.Receive(outPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	}))

	signal <- uuid.Must(uuid.NewV7())

	select {
	case <-done:
	case <-ctx.Done():
		require.Fail(t, ctx.Err().Error())
	}
}
