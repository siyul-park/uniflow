package system

import (
	"context"
	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestNewSignalNode(t *testing.T) {
	n := NewSignalNode(nil)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestSignalNode_Port(t *testing.T) {
	n := NewSignalNode(nil)
	defer n.Close()

	assert.NotNil(t, n.Out(node.PortOut))
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
			assert.Fail(t, ctx.Err().Error())
		}
	}))

	signal <- uuid.Must(uuid.NewV4())

	select {
	case <-done:
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}
