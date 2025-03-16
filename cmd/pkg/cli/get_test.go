package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/meta"
	"github.com/siyul-park/uniflow/pkg/store"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/stretchr/testify/require"
)

func TestGetCommand_Execute(t *testing.T) {
	specStore := store.New()
	valueStore := store.New()

	t.Run("GetSpec", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		kind := faker.UUIDHyphenated()

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: meta.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		}

		err := specStore.Insert(ctx, []any{meta})
		require.NoError(t, err)

		output := new(bytes.Buffer)

		cmd := NewGetCommand(GetConfig{
			SpecStore:  specStore,
			ValueStore: valueStore,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{specs})

		err = cmd.Execute()
		require.NoError(t, err)
		require.Contains(t, output.String(), meta.Name)
	})

	t.Run("GetValue", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		val := &value.Value{
			ID:        uuid.Must(uuid.NewV7()),
			Namespace: meta.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Data:      faker.UUIDHyphenated(),
		}

		err := valueStore.Insert(ctx, []any{val})
		require.NoError(t, err)

		output := new(bytes.Buffer)

		cmd := NewGetCommand(GetConfig{
			SpecStore:  specStore,
			ValueStore: valueStore,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{values})

		err = cmd.Execute()
		require.NoError(t, err)
		require.Contains(t, output.String(), val.Name)
	})
}
