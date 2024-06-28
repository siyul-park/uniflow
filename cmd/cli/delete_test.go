package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"testing/fstest"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/database/memdb"
	"github.com/siyul-park/uniflow/node"
	"github.com/siyul-park/uniflow/scheme"
	"github.com/siyul-park/uniflow/spec"
	"github.com/siyul-park/uniflow/store"
	"github.com/stretchr/testify/assert"
)

func TestDeleteCommand_Execute(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := scheme.New()
	db := memdb.New("")
	fsys := make(fstest.MapFS)

	st, _ := store.New(ctx, store.Config{
		Scheme:   s,
		Database: db,
	})

	kind := faker.UUIDHyphenated()

	codec := scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, codec)

	filename := "patch.json"

	meta := &spec.Meta{
		Kind:      kind,
		Namespace: spec.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	data, _ := json.Marshal(meta)

	fsys[filename] = &fstest.MapFile{
		Data: data,
	}

	_, _ = st.InsertOne(ctx, meta)

	cmd := NewDeleteCommand(DeleteConfig{
		Scheme:   s,
		Database: db,
		FS:       fsys,
	})

	cmd.SetArgs([]string{fmt.Sprintf("--%s", flagFilename), filename})

	err := cmd.Execute()
	assert.NoError(t, err)

	r, err := st.FindOne(ctx, store.Where[string](spec.KeyName).EQ(meta.GetName()))
	assert.NoError(t, err)
	assert.Nil(t, r)
}
