package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/pkg/driver"
	"github.com/siyul-park/uniflow/pkg/meta"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/value"
)

func TestDeleteCommand_Execute(t *testing.T) {
	specStore := driver.NewStore()
	valueStore := driver.NewStore()

	fs := afero.NewMemMapFs()

	t.Run("DeleteSpec", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "specs.json"

		kind := faker.UUIDHyphenated()

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: meta.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		}

		data, err := json.Marshal(meta)
		require.NoError(t, err)

		file, err := fs.Create(filename)
		require.NoError(t, err)
		defer file.Close()

		_, err = file.Write(data)
		require.NoError(t, err)

		err = specStore.Insert(ctx, []any{meta})
		require.NoError(t, err)

		cmd := NewDeleteCommand(DeleteConfig{
			SpecStore:  specStore,
			ValueStore: valueStore,
			FS:         fs,
		})

		cmd.SetArgs([]string{specs, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		require.NoError(t, err)

		cursor, err := specStore.Find(ctx, meta)
		require.NoError(t, err)
		require.False(t, cursor.Next(ctx))
	})

	t.Run("DeleteValue", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "values.json"

		val := &value.Value{
			ID:        uuid.Must(uuid.NewV7()),
			Namespace: meta.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Data:      faker.UUIDHyphenated(),
		}

		data, err := json.Marshal(val)
		require.NoError(t, err)

		file, err := fs.Create(filename)
		require.NoError(t, err)
		defer file.Close()

		_, err = file.Write(data)
		require.NoError(t, err)

		err = valueStore.Insert(ctx, []any{val})
		require.NoError(t, err)

		cmd := NewDeleteCommand(DeleteConfig{
			SpecStore:  specStore,
			ValueStore: valueStore,
			FS:         fs,
		})

		cmd.SetArgs([]string{values, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		require.NoError(t, err)

		cursor, err := valueStore.Find(ctx, val)
		require.NoError(t, err)
		require.False(t, cursor.Next(ctx))
	})
}
