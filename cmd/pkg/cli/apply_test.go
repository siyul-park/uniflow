package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestApplyCommand_Execute(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	st := spec.NewStore()
	fsys := afero.NewMemMapFs()

	kind := faker.UUIDHyphenated()

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
		SpecStore: st,
		FS:        fsys,
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
