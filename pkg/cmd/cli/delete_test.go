package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"testing/fstest"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestDeleteCommand_Execute(t *testing.T) {
	s := scheme.New()
	db := memdb.New("")
	fsys := make(fstest.MapFS)

	st, _ := storage.New(context.Background(), storage.Config{
		Scheme:   s,
		Database: db,
	})

	kind := faker.UUIDHyphenated()

	codec := scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, codec)

	filename := "patch.json"

	spec := &scheme.SpecMeta{
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	data, _ := json.Marshal(spec)

	fsys[filename] = &fstest.MapFile{
		Data: data,
	}

	_, _ = st.InsertOne(context.Background(), spec)

	cmd := NewDeleteCommand(DeleteConfig{
		Scheme:   s,
		Database: db,
		FS:       fsys,
	})

	cmd.SetArgs([]string{fmt.Sprintf("--%s", flagFilename), filename})

	err := cmd.Execute()
	assert.NoError(t, err)

	r, err := st.FindOne(context.Background(), storage.Where[string](scheme.KeyName).EQ(spec.GetName()))
	assert.NoError(t, err)
	assert.Nil(t, r)
}
