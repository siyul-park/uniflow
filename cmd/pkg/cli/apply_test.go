package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestApplyCommand_Execute(t *testing.T) {
	specStore := store.New()
	valueStore := store.New()

	fs := afero.NewMemMapFs()

	t.Run("InsertSpec", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "specs.json"

		kind := faker.UUIDHyphenated()

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		}

		data, err := json.Marshal(meta)
		require.NoError(t, err)

		file, err := fs.Create(filename)
		require.NoError(t, err)
		defer file.Close()

		_, err = file.Write(data)
		require.NoError(t, err)

		output := new(bytes.Buffer)

		cmd := NewApplyCommand(ApplyConfig{
			SpecStore:  specStore,
			ValueStore: valueStore,
			FS:         fs,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{specs, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		require.NoError(t, err)

		cursor, err := specStore.Find(ctx, nil)
		require.NoError(t, err)
		require.True(t, cursor.Next(ctx))
		require.Contains(t, output.String(), meta.Name)
	})

	t.Run("UpdateSpec", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "specs.json"

		kind := faker.UUIDHyphenated()

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		}

		err := specStore.Insert(ctx, []any{meta})
		require.NoError(t, err)

		data, err := json.Marshal(meta)
		require.NoError(t, err)

		file, err := fs.Create(filename)
		require.NoError(t, err)
		defer file.Close()

		_, err = file.Write(data)
		require.NoError(t, err)

		output := new(bytes.Buffer)

		cmd := NewApplyCommand(ApplyConfig{
			SpecStore:  specStore,
			ValueStore: valueStore,
			FS:         fs,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{specs, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		require.NoError(t, err)

		cursor, err := specStore.Find(ctx, nil)
		require.NoError(t, err)
		require.True(t, cursor.Next(ctx))
		require.Contains(t, output.String(), meta.Name)
	})

	t.Run("InsertValue", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "values.json"

		val := &value.Value{
			ID:        uuid.Must(uuid.NewV7()),
			Namespace: resource.DefaultNamespace,
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

		output := new(bytes.Buffer)

		cmd := NewApplyCommand(ApplyConfig{
			SpecStore:  specStore,
			ValueStore: valueStore,
			FS:         fs,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{values, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		require.NoError(t, err)

		cursor, err := valueStore.Find(ctx, nil)
		require.NoError(t, err)
		require.True(t, cursor.Next(ctx))
		require.Contains(t, output.String(), val.Name)
	})

	t.Run("UpdateValue", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "values.json"

		val := &value.Value{
			ID:        uuid.Must(uuid.NewV7()),
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Data:      faker.UUIDHyphenated(),
		}

		err := valueStore.Insert(ctx, []any{val})
		require.NoError(t, err)

		data, err := json.Marshal(val)
		require.NoError(t, err)

		file, err := fs.Create(filename)
		require.NoError(t, err)
		defer file.Close()

		_, err = file.Write(data)
		require.NoError(t, err)

		output := new(bytes.Buffer)

		cmd := NewApplyCommand(ApplyConfig{
			SpecStore:  specStore,
			ValueStore: valueStore,
			FS:         fs,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{values, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		require.NoError(t, err)

		cursor, err := valueStore.Find(ctx, nil)
		require.NoError(t, err)
		require.True(t, cursor.Next(ctx))
		require.Contains(t, output.String(), val.Name)
	})
}
