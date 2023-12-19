package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"testing/fstest"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestApplyCommand_Execute(t *testing.T) {
	s := scheme.New()
	db := memdb.New("")
	fsys := make(fstest.MapFS)

	st, _ := storage.New(context.Background(), storage.Config{
		Scheme:   s,
		Database: db,
	})

	patchFilepath := "patch.json"
	kind := faker.Word()

	spec := &scheme.SpecMeta{
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
		Name:      faker.Word(),
	}

	codec := scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, codec)

	data, _ := json.Marshal(spec)

	fsys[patchFilepath] = &fstest.MapFile{
		Data: data,
	}

	output := new(bytes.Buffer)

	cmd := NewApplyCommand(ApplyConfig{
		Scheme:   s,
		Database: db,
		FS:       fsys,
	})
	cmd.SetOut(output)
	cmd.SetErr(output)

	cmd.SetArgs([]string{"--file", patchFilepath})

	err := cmd.Execute()
	assert.NoError(t, err)

	r, err := st.FindOne(context.Background(), storage.Where[string](scheme.KeyName).EQ(spec.GetName()))
	assert.NoError(t, err)
	assert.NotNil(t, r)

	assert.Contains(t, output.String(), spec.Name)
}
