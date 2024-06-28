package control

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/node"
	"github.com/siyul-park/uniflow/object"
	"github.com/siyul-park/uniflow/packet"
	"github.com/siyul-park/uniflow/port"
	"github.com/siyul-park/uniflow/process"
	"github.com/stretchr/testify/assert"
)

func TestNewMergeNode(t *testing.T) {
	n := NewMergeNode()
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestMergeNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	n := NewMergeNode()
	defer n.Close()

	var ins []*port.OutPort
	for i := 0; i < 4; i++ {
		in := port.NewOut()
		in.Link(n.In(node.PortWithIndex(node.PortIn, i)))
		ins = append(ins, in)
	}

	out := port.NewIn()
	n.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	inWriters := make([]*packet.Writer, len(ins))
	for i, in := range ins {
		inWriters[i] = in.Open(proc)
	}
	outReader := out.Open(proc)

	var inPayloads []object.Object
	for range inWriters {
		inPayloads = append(inPayloads, object.NewString(faker.UUIDHyphenated()))
	}

	merged := object.NewSlice(inPayloads...).Interface()

	for i, inWriter := range inWriters {
		inPck := packet.New(inPayloads[i])
		inWriter.Write(inPck)
	}

	select {
	case outPck := <-outReader.Read():
		assert.Equal(t, merged, outPck.Payload().Interface())
		outReader.Receive(outPck)
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}

	for _, inWriter := range inWriters {
		select {
		case backPck := <-inWriter.Receive():
			assert.NotNil(t, backPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	}
}

func TestMergeNodeCodec_Decode(t *testing.T) {
	codec := NewMergeNodeCodec()

	spec := &MergeNodeSpec{}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func BenchmarkMergeNode_SendAndReceive(b *testing.B) {
	n := NewMergeNode()
	defer n.Close()

	var ins []*port.OutPort
	for i := 0; i < 4; i++ {
		in := port.NewOut()
		in.Link(n.In(node.PortWithIndex(node.PortIn, i)))
		ins = append(ins, in)
	}

	out := port.NewIn()
	n.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	inWriters := make([]*packet.Writer, len(ins))
	for i, in := range ins {
		inWriters[i] = in.Open(proc)
	}
	outReader := out.Open(proc)

	var inPayloads []object.Object
	for range inWriters {
		inPayloads = append(inPayloads, object.NewString(faker.UUIDHyphenated()))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for i, inWriter := range inWriters {
			inPck := packet.New(inPayloads[i])
			inWriter.Write(inPck)
		}

		outPck := <-outReader.Read()
		outReader.Receive(outPck)

		for _, inWriter := range inWriters {
			<-inWriter.Receive()
		}
	}
}
