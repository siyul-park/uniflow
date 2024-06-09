package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/assert"
)

func TestGetCommand_Execute(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := scheme.New()
	db := memdb.New("")

	st, _ := scheme.NewStorage(ctx, scheme.StorageConfig{
		Scheme:   s,
		Database: db,
	})

	kind := faker.UUIDHyphenated()

	codec := scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, codec)

	spec := &scheme.SpecMeta{
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	id, _ := st.InsertOne(ctx, spec)

	output := new(bytes.Buffer)

	cmd := NewGetCommand(GetConfig{
		Scheme:   s,
		Database: db,
	})
	cmd.SetOut(output)
	cmd.SetErr(output)

	err := cmd.Execute()
	assert.NoError(t, err)

	assert.Contains(t, output.String(), id.String())
}
