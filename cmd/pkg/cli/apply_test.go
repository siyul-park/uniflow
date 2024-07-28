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
	specStore := spec.NewStore()
	secretStore := secret.NewStore()
	fs := afero.NewMemMapFs()

	t.Run("Apply Node Spec", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "nodes.json"

		kind := faker.UUIDHyphenated()

		meta := &spec.Meta{
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		}

		dataMeta, err := json.Marshal(meta)
		assert.NoError(t, err)

		fileMeta, err := fs.Create(filename)
		assert.NoError(t, err)
		defer fileMeta.Close()

		_, err = fileMeta.Write(dataMeta)
		assert.NoError(t, err)

		output := new(bytes.Buffer)

		cmd := NewApplyCommand(ApplyConfig{
			SpecStore:   specStore,
			SecretStore: secretStore,
			FS:          fs,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{argNodes, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		assert.NoError(t, err)

		results, err := specStore.Load(ctx, meta)
		assert.NoError(t, err)
		assert.Len(t, results, 1)

		assert.Contains(t, output.String(), meta.Name)
	})

	t.Run("Apply Secret", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "secrets.json"

		sec := &secret.Secret{
			Namespace: spec.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		}

		dataSecret, err := json.Marshal(sec)
		assert.NoError(t, err)

		fileSecret, err := fs.Create(filename)
		assert.NoError(t, err)
		defer fileSecret.Close()

		_, err = fileSecret.Write(dataSecret)
		assert.NoError(t, err)

		output := new(bytes.Buffer)

		cmd := NewApplyCommand(ApplyConfig{
			SpecStore:   specStore,
			SecretStore: secretStore,
			FS:          fs,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{argSecrets, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		assert.NoError(t, err)

		results, err := secretStore.Load(ctx, sec)
		assert.NoError(t, err)
		assert.Len(t, results, 1)

		assert.Contains(t, output.String(), sec.Name)
	})
}
