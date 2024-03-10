package system

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestGetNodes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := storage.New(ctx, storage.Config{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	n, _ := NewBridgeNode(GetNodes(st))
	defer n.Close()

	spec := &scheme.SpecMeta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	id, _ := st.InsertOne(ctx, spec)

	io := port.NewOut()
	io.Link(n.In(node.PortIO))

	proc := process.New()
	defer proc.Close()

	ioWriter := io.Open(proc)

	inPayload, _ := primitive.MarshalBinary(storage.Where[uuid.UUID]("id").EQ(id))
	inPck := packet.New(inPayload)

	ioWriter.Write(inPck)

	select {
	case outPck := <-ioWriter.Receive():
		var outPayload []*scheme.SpecMeta
		assert.NoError(t, primitive.Unmarshal(outPck.Payload(), &outPayload))
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}
