package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/stretchr/testify/require"
)

func TestGetCommand_Execute(t *testing.T) {
	specStore := spec.NewStore()
	valueStore := value.NewStore()

	t.Run("GetSpec", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		kind := faker.UUIDHyphenated()

		meta := &spec.Meta{
			Kind: kind,
			Name: faker.UUIDHyphenated(),
		}

		_, err := specStore.Store(ctx, meta)
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

		scrt := &value.Value{
			Name: faker.UUIDHyphenated(),
			Data: faker.UUIDHyphenated(),
		}

		_, err := valueStore.Store(ctx, scrt)
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

		require.Contains(t, output.String(), scrt.Name)
	})
}
