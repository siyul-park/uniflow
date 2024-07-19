package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestApplyCommand_Execute(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := scheme.New()
	st := spec.NewStore(memdb.NewCollection(""))
	fsys := afero.NewMemMapFs()

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

	f, _ := fsys.Create(filename)
	f.Write(data)

	output := new(bytes.Buffer)

	cmd := NewApplyCommand(ApplyConfig{
		Scheme: s,
		Store:  st,
		FS:     fsys,
	})
	cmd.SetOut(output)
	cmd.SetErr(output)

	cmd.SetArgs([]string{fmt.Sprintf("--%s", flagFilename), filename})

	err := cmd.Execute()
	assert.NoError(t, err)

	r, err := st.Load(ctx, meta)
	assert.NoError(t, err)
	assert.Len(t, r, 1)

	assert.Contains(t, output.String(), meta.Name)
}
