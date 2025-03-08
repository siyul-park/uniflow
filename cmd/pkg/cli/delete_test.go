package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestDeleteCommand_Execute(t *testing.T) {
	specStore := spec.NewStore()
	valueStore := value.NewStore()

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
		require.NoError(t, err)

		file, err := fs.Create(filename)
		require.NoError(t, err)
		defer file.Close()

		_, err = file.Write(data)
		require.NoError(t, err)

		_, err = specStore.Store(ctx, meta)
		require.NoError(t, err)

		cmd := NewDeleteCommand(DeleteConfig{
			SpecStore:  specStore,
			ValueStore: valueStore,
			FS:         fs,
		})

		cmd.SetArgs([]string{specs, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		require.NoError(t, err)

		r, err := specStore.Load(ctx, meta)
		require.NoError(t, err)
		require.Len(t, r, 0)
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
		require.NoError(t, err)

		file, err := fs.Create(filename)
		require.NoError(t, err)
		defer file.Close()

		_, err = file.Write(data)
		require.NoError(t, err)

		_, err = valueStore.Store(ctx, scrt)
		require.NoError(t, err)

		cmd := NewDeleteCommand(DeleteConfig{
			SpecStore:  specStore,
			ValueStore: valueStore,
			FS:         fs,
		})

		cmd.SetArgs([]string{values, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		require.NoError(t, err)

		rValue, err := valueStore.Load(ctx, scrt)
		require.NoError(t, err)
		require.Len(t, rValue, 0)
	})
}
