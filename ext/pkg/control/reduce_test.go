package control

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/ext/pkg/language/text"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestReduceNodeCodec_Decode(t *testing.T) {
	codec := NewReduceNodeCodec(text.NewCompiler())

	spec := &ReduceNodeSpec{}

	n, err := codec.Compile(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestNewReduceNode(t *testing.T) {
	n := NewReduceNode(func(a1, a2 any, i int) (any, error) {
		return a1, nil
	}, nil)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestReduceNode_Port(t *testing.T) {
	n := NewReduceNode(func(a1, a2 any, i int) (any, error) {
		return a1, nil
	}, nil)
	defer n.Close()

	assert.NotNil(t, n.In(node.PortIn))
	assert.NotNil(t, n.Out(node.PortOut))
	assert.NotNil(t, n.Out(node.PortErr))
}

func TestReduceNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	n := NewReduceNode(func(a1, a2 any, i int) (any, error) {
		return a2, nil
	}, nil)
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	out := port.NewIn()
	n.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	outReader := out.Open(proc)

	inPayload := types.NewString("foo")
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-outReader.Read():
		assert.Equal(t, inPayload, outPck.Payload())
		outReader.Receive(outPck)
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}

	select {
	case backPck := <-inWriter.Receive():
		assert.NotNil(t, backPck)
	case <-ctx.Done():
		assert.Fail(t, "timeout")
	}

	inPayload = types.NewString("bar")
	inPck = packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-outReader.Read():
		assert.Equal(t, inPayload, outPck.Payload())
		outReader.Receive(outPck)
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}

	select {
	case backPck := <-inWriter.Receive():
		assert.NotNil(t, backPck)
	case <-ctx.Done():
		assert.Fail(t, "timeout")
	}
}

func BenchmarkReduceNode_SendAndReceive(b *testing.B) {
	n := NewReduceNode(func(a1, a2 any, i int) (any, error) {
		return a2, nil
	}, nil)
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	out := port.NewIn()
	n.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	outReader := out.Open(proc)

	inPayload := types.NewString("foo")
	inPck := packet.New(inPayload)

	b.ResetTimer()

	inWriter.Write(inPck)

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)

		outPck := <-outReader.Read()
		outReader.Receive(outPck)

		<-inWriter.Receive()
	}
}
