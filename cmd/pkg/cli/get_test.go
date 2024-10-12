package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/chart"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestGetCommand_Execute(t *testing.T) {
	specStore := spec.NewStore()
	secretStore := secret.NewStore()
	chartStore := chart.NewStore()

	t.Run("GetSpec", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		kind := faker.UUIDHyphenated()

		meta := &spec.Meta{
			Kind: kind,
			Name: faker.UUIDHyphenated(),
		}

		_, err := specStore.Store(ctx, meta)
		assert.NoError(t, err)

		output := new(bytes.Buffer)

		cmd := NewGetCommand(GetConfig{
			SpecStore:   specStore,
			SecretStore: secretStore,
			ChartStore:  chartStore,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{specs})

		err = cmd.Execute()
		assert.NoError(t, err)

		assert.Contains(t, output.String(), meta.Name)
	})

	t.Run("GetSecret", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		scrt := &secret.Secret{
			Name: faker.UUIDHyphenated(),
			Data: faker.Word(),
		}

		_, err := secretStore.Store(ctx, scrt)
		assert.NoError(t, err)

		output := new(bytes.Buffer)

		cmd := NewGetCommand(GetConfig{
			SpecStore:   specStore,
			SecretStore: secretStore,
			ChartStore:  chartStore,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{secrets})

		err = cmd.Execute()
		assert.NoError(t, err)

		assert.Contains(t, output.String(), scrt.Name)
	})

	t.Run("GetChart", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		chrt := &chart.Chart{
			ID:   uuid.Must(uuid.NewV7()),
			Name: faker.Word(),
		}

		_, err := chartStore.Store(ctx, chrt)
		assert.NoError(t, err)

		output := new(bytes.Buffer)

		cmd := NewGetCommand(GetConfig{
			SpecStore:   specStore,
			SecretStore: secretStore,
			ChartStore:  chartStore,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{charts})

		err = cmd.Execute()
		assert.NoError(t, err)

		assert.Contains(t, output.String(), chrt.Name)
	})
}
