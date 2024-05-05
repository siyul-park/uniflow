package control

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewNoOpNode(t *testing.T) {
	n := NewNoOpNode()
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestNoOpNode_Port(t *testing.T) {
	n := NewNoOpNode()
	defer n.Close()

	assert.NotNil(t, n.In(node.PortIn))
}

func TestNoOpNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	n := NewNoOpNode()
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
		assert.Fail(t, "timeout")
	}
}

func TestNoOpNodeCodec_Decode(t *testing.T) {
	codec := NewNoOpNodeCodec()

	spec := &NoOpNodeSpec{}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}
