package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestGetCommand_Execute(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	specStore := spec.NewStore()
	secretStore := secret.NewStore()

	t.Run("Get Node Spec", func(t *testing.T) {
		kind := faker.UUIDHyphenated()

		meta := &spec.Meta{
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		}

		_, err := specStore.Store(ctx, meta)
		assert.NoError(t, err)

		output := new(bytes.Buffer)

		cmd := NewGetCommand(GetConfig{
			SpecStore:   specStore,
			SecretStore: secretStore,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{argNodes})

		err = cmd.Execute()
		assert.NoError(t, err)

		assert.Contains(t, output.String(), meta.Name)
	})

	t.Run("Get Secret", func(t *testing.T) {
		sec := &secret.Secret{
			Namespace: spec.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		}

		_, err := secretStore.Store(ctx, sec)
		assert.NoError(t, err)

		output := new(bytes.Buffer)

		cmd := NewGetCommand(GetConfig{
			SpecStore:   specStore,
			SecretStore: secretStore,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{argSecrets})

		err = cmd.Execute()
		assert.NoError(t, err)

		assert.Contains(t, output.String(), sec.Name)
	})
}
