package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestApplyCommand_Execute(t *testing.T) {
	t.Run("nodes", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		spst := spec.NewStore()
		scst := secret.NewStore()

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
			SpecStore:   spst,
			SecretStore: scst,
			FS:          fsys,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)

		cmd.SetArgs([]string{argNodes, fmt.Sprintf("--%s", flagFilename), filename})

		err := cmd.Execute()
		assert.NoError(t, err)

		r, err := spst.Load(ctx, meta)
		assert.NoError(t, err)
		assert.Len(t, r, 1)

		assert.Contains(t, output.String(), meta.Name)
	})

	t.Run("secrets", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		spst := spec.NewStore()
		scst := secret.NewStore()

		fsys := afero.NewMemMapFs()

		filename := "patch.json"

		secret := &secret.Secret{
			Namespace: spec.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		}

		data, _ := json.Marshal(secret)

		f, _ := fsys.Create(filename)
		f.Write(data)

		output := new(bytes.Buffer)

		cmd := NewApplyCommand(ApplyConfig{
			SpecStore:   spst,
			SecretStore: scst,
			FS:          fsys,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)

		cmd.SetArgs([]string{argSecrets, fmt.Sprintf("--%s", flagFilename), filename})

		err := cmd.Execute()
		assert.NoError(t, err)

		r, err := scst.Load(ctx, secret)
		assert.NoError(t, err)
		assert.Len(t, r, 1)

		assert.Contains(t, output.String(), secret.Name)
	})
}
