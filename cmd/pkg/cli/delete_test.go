package cli

import (
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

func TestDeleteCommand_Execute(t *testing.T) {
	specStore := spec.NewStore()
	secretStore := secret.NewStore()
	fs := afero.NewMemMapFs()

	t.Run("Delete Node Spec", func(t *testing.T) {
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

		_, err = specStore.Store(ctx, meta)
		assert.NoError(t, err)

		cmd := NewDeleteCommand(DeleteConfig{
			SpecStore:   specStore,
			SecretStore: secretStore,
			FS:          fs,
		})

		cmd.SetArgs([]string{argNodes, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		assert.NoError(t, err)

		r, err := specStore.Load(ctx, meta)
		assert.NoError(t, err)
		assert.Len(t, r, 0)
	})

	t.Run("Delete Secret", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "secrets.json"

		secret := &secret.Secret{
			Namespace: spec.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		}

		dataSecret, err := json.Marshal(secret)
		assert.NoError(t, err)

		fileSecret, err := fs.Create(filename)
		assert.NoError(t, err)
		defer fileSecret.Close()

		_, err = fileSecret.Write(dataSecret)
		assert.NoError(t, err)

		_, err = secretStore.Store(ctx, secret)
		assert.NoError(t, err)

		cmd := NewDeleteCommand(DeleteConfig{
			SpecStore:   specStore,
			SecretStore: secretStore,
			FS:          fs,
		})

		cmd.SetArgs([]string{argSecrets, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		assert.NoError(t, err)

		rSecret, err := secretStore.Load(ctx, secret)
		assert.NoError(t, err)
		assert.Len(t, rSecret, 0)
	})
}
