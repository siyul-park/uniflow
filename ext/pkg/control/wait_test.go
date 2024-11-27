package control

import (
	"context"
	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestWaitNodeCodec_Compile(t *testing.T) {
	codec := NewWaitNodeCodec()

	spec := &WaitNodeSpec{
		Interval: 0,
	}

	n, err := codec.Compile(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestNewWaitNode(t *testing.T) {
	n := NewWaitNode(0)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestWaitNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	n := NewWaitNode(0)
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := types.NewString(faker.Word())
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		assert.Equal(t, inPayload, outPck.Payload())
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func BenchmarkWaitNode_SendAndReceive(b *testing.B) {
	n := NewWaitNode(0)
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	var inPayload types.Value
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		<-inWriter.Receive()
	}
}
