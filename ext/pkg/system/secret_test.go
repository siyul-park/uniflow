package system

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestCreateSecrets(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	st := secret.NewStore()

	n, _ := NewSyscallNode(CreateSecrets(st))
	defer n.Close()

	sec := &secret.Secret{
		ID: uuid.Must(uuid.NewV7()),
		Data: faker.Word(),
	}

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload, _ := types.Encoder.Encode(sec)
	inPck := packet.New(types.NewSlice(inPayload))

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		var outPayload []*secret.Secret
		assert.NoError(t, types.Decoder.Decode(outPck.Payload(), &outPayload))
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func TestReadSecrets(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	st := secret.NewStore()

	n, _ := NewSyscallNode(ReadSecrets(st))
	defer n.Close()

	sec := &secret.Secret{
		ID: uuid.Must(uuid.NewV7()),
		Data: faker.Word(),
	}

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload, _ := types.Encoder.Encode(sec)
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		var outPayload []*secret.Secret
		assert.NoError(t, types.Decoder.Decode(outPck.Payload(), &outPayload))
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func TestUpdateSecrets(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	st := secret.NewStore()

	n, _ := NewSyscallNode(UpdateSecrets(st))
	defer n.Close()

	sec := &secret.Secret{
		ID: uuid.Must(uuid.NewV7()),
		Data: faker.Word(),
	}

	_, _ = st.Store(ctx, sec)

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload, _ := types.Encoder.Encode(sec)
	inPck := packet.New(types.NewSlice(inPayload))

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		var outPayload []*secret.Secret
		assert.NoError(t, types.Decoder.Decode(outPck.Payload(), &outPayload))
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func TestDeleteSecrets(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	st := secret.NewStore()

	n, _ := NewSyscallNode(DeleteSecrets(st))
	defer n.Close()

	sec := &secret.Secret{
		ID: uuid.Must(uuid.NewV7()),
		Data: faker.Word(),
	}

	_, _ = st.Store(ctx, sec)

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload, _ := types.Encoder.Encode(sec)
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		var outPayload []*secret.Secret
		assert.NoError(t, types.Decoder.Decode(outPck.Payload(), &outPayload))
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}
