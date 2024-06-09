package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"testing/fstest"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestApplyCommand_Execute(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := spec.NewScheme()
	db := memdb.New("")
	fsys := make(fstest.MapFS)

	st, _ := spec.NewStorage(ctx, spec.StorageConfig{
		Scheme:   s,
		Database: db,
	})

	kind := faker.UUIDHyphenated()

	codec := spec.CodecFunc(func(spec spec.Spec) (node.Node, error) {
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

	output := new(bytes.Buffer)

	cmd := NewApplyCommand(ApplyConfig{
		Scheme:   s,
		Database: db,
		FS:       fsys,
	})
	cmd.SetOut(output)
	cmd.SetErr(output)

	cmd.SetArgs([]string{fmt.Sprintf("--%s", flagFilename), filename})

	err := cmd.Execute()
	assert.NoError(t, err)

	r, err := st.FindOne(ctx, spec.Where[string](spec.KeyName).EQ(meta.GetName()))
	assert.NoError(t, err)
	assert.NotNil(t, r)

	assert.Contains(t, output.String(), meta.Name)
}
