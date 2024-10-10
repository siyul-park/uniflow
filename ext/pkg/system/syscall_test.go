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
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestCreateResource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	st := resource.NewStore[*resource.Meta]()

	n, _ := NewNativeNode(CreateResource(st))
	defer n.Close()

	meta := &resource.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Name: faker.Word(),
	}

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload, _ := types.Marshal(meta)
	inPck := packet.New(types.NewSlice(inPayload))

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		var outPayload []*resource.Meta
		assert.NoError(t, types.Unmarshal(outPck.Payload(), &outPayload))
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func TestReadResource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	st := resource.NewStore[*resource.Meta]()

	n, _ := NewNativeNode(ReadResource(st))
	defer n.Close()

	meta := &resource.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Name: faker.Word(),
	}

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload, _ := types.Marshal(meta)
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		var outPayload []*resource.Meta
		assert.NoError(t, types.Unmarshal(outPck.Payload(), &outPayload))
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func TestUpdateResource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	st := resource.NewStore[*resource.Meta]()

	n, _ := NewNativeNode(UpdateResource(st))
	defer n.Close()

	meta := &resource.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Name: faker.Word(),
	}

	_, _ = st.Store(ctx, meta)

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload, _ := types.Marshal(meta)
	inPck := packet.New(types.NewSlice(inPayload))

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		var outPayload []*resource.Meta
		assert.NoError(t, types.Unmarshal(outPck.Payload(), &outPayload))
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func TestDeleteResource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	st := resource.NewStore[*resource.Meta]()

	n, _ := NewNativeNode(DeleteResource(st))
	defer n.Close()

	meta := &resource.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Name: faker.Word(),
	}

	_, _ = st.Store(ctx, meta)

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload, _ := types.Marshal(meta)
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		var outPayload []*resource.Meta
		assert.NoError(t, types.Unmarshal(outPck.Payload(), &outPayload))
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}
