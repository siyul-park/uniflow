package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/chart"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestDeleteCommand_Execute(t *testing.T) {
	chartStore := chart.NewStore()
	specStore := spec.NewStore()
	secretStore := secret.NewStore()

	fs := afero.NewMemMapFs()

	t.Run("DeleteChart", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "chart.json"

		chrt := &chart.Chart{
			ID:   uuid.Must(uuid.NewV7()),
			Name: faker.Word(),
		}

		data, err := json.Marshal(chrt)
		assert.NoError(t, err)

		file, err := fs.Create(filename)
		assert.NoError(t, err)
		defer file.Close()

		_, err = file.Write(data)
		assert.NoError(t, err)

		_, err = chartStore.Store(ctx, chrt)
		assert.NoError(t, err)

		cmd := NewDeleteCommand(DeleteConfig{
			ChartStore:  chartStore,
			SpecStore:   specStore,
			SecretStore: secretStore,
			FS:          fs,
		})

		cmd.SetArgs([]string{argCharts, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		assert.NoError(t, err)

		r, err := chartStore.Load(ctx, chrt)
		assert.NoError(t, err)
		assert.Len(t, r, 0)
	})

	t.Run("DeleteNodeSpec", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "nodes.json"

		kind := faker.UUIDHyphenated()

		meta := &spec.Meta{
			Kind: kind,
			Name: faker.UUIDHyphenated(),
		}

		data, err := json.Marshal(meta)
		assert.NoError(t, err)

		file, err := fs.Create(filename)
		assert.NoError(t, err)
		defer file.Close()

		_, err = file.Write(data)
		assert.NoError(t, err)

		_, err = specStore.Store(ctx, meta)
		assert.NoError(t, err)

		cmd := NewDeleteCommand(DeleteConfig{
			ChartStore:  chartStore,
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

	t.Run("DeleteSecret", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "secrets.json"

		scrt := &secret.Secret{
			Name: faker.UUIDHyphenated(),
			Data: faker.Word(),
		}

		data, err := json.Marshal(scrt)
		assert.NoError(t, err)

		file, err := fs.Create(filename)
		assert.NoError(t, err)
		defer file.Close()

		_, err = file.Write(data)
		assert.NoError(t, err)

		_, err = secretStore.Store(ctx, scrt)
		assert.NoError(t, err)

		cmd := NewDeleteCommand(DeleteConfig{
			ChartStore:  chartStore,
			SpecStore:   specStore,
			SecretStore: secretStore,
			FS:          fs,
		})

		cmd.SetArgs([]string{argSecrets, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		assert.NoError(t, err)

		rSecret, err := secretStore.Load(ctx, scrt)
		assert.NoError(t, err)
		assert.Len(t, rSecret, 0)
	})
}
