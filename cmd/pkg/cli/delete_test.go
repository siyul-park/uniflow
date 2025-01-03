package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/chart"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestDeleteCommand_Execute(t *testing.T) {
	specStore := spec.NewStore()
	valueStore := value.NewStore()
	chartStore := chart.NewStore()

	fs := afero.NewMemMapFs()

	t.Run("DeleteSpec", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "specs.json"

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
			SpecStore:  specStore,
			ValueStore: valueStore,
			ChartStore: chartStore,
			FS:         fs,
		})

		cmd.SetArgs([]string{specs, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		assert.NoError(t, err)

		r, err := specStore.Load(ctx, meta)
		assert.NoError(t, err)
		assert.Len(t, r, 0)
	})

	t.Run("DeleteValue", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "values.json"

		scrt := &value.Value{
			Name: faker.UUIDHyphenated(),
			Data: faker.UUIDHyphenated(),
		}

		data, err := json.Marshal(scrt)
		assert.NoError(t, err)

		file, err := fs.Create(filename)
		assert.NoError(t, err)
		defer file.Close()

		_, err = file.Write(data)
		assert.NoError(t, err)

		_, err = valueStore.Store(ctx, scrt)
		assert.NoError(t, err)

		cmd := NewDeleteCommand(DeleteConfig{
			SpecStore:  specStore,
			ValueStore: valueStore,
			ChartStore: chartStore,
			FS:         fs,
		})

		cmd.SetArgs([]string{values, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		assert.NoError(t, err)

		rValue, err := valueStore.Load(ctx, scrt)
		assert.NoError(t, err)
		assert.Len(t, rValue, 0)
	})

	t.Run("DeleteChart", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "charts.json"

		chrt := &chart.Chart{
			ID:   uuid.Must(uuid.NewV7()),
			Name: faker.UUIDHyphenated(),
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
			SpecStore:  specStore,
			ValueStore: valueStore,
			ChartStore: chartStore,
			FS:         fs,
		})

		cmd.SetArgs([]string{charts, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		assert.NoError(t, err)

		r, err := chartStore.Load(ctx, chrt)
		assert.NoError(t, err)
		assert.Len(t, r, 0)
	})
}
