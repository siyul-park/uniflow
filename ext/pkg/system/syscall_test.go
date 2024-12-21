package system

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestSyscallNodeCodec_Compile(t *testing.T) {
	opcode := faker.UUIDHyphenated()

	codec := NewSyscallNodeCodec(map[string]func(ctx context.Context, arguments []any) ([]any, error){
		opcode: func(ctx context.Context, arguments []any) ([]any, error) {
			return nil, nil
		},
	})

	spec := &SyscallNodeSpec{
		OPCode: opcode,
	}

	n, err := codec.Compile(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestNewSyscallNode(t *testing.T) {
	n, err := NewSyscallNode(nil)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestSyscallNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	n, _ := NewSyscallNode(func(ctx context.Context, arguments []any) ([]any, error) {
		return arguments, nil
	})
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := types.NewString(faker.UUIDHyphenated())
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		assert.Equal(t, inPayload, outPck.Payload())
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func BenchmarkSyscallNode_SendAndReceive(b *testing.B) {
	n, _ := NewSyscallNode(func(ctx context.Context, arguments []any) ([]any, error) {
		return arguments, nil
	})
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := types.NewString(faker.UUIDHyphenated())
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		<-inWriter.Receive()
	}
}
