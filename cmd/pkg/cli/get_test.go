package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/stretchr/testify/assert"
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
		assert.NoError(t, err)

		output := new(bytes.Buffer)

		cmd := NewGetCommand(GetConfig{
			SpecStore:  specStore,
			ValueStore: valueStore,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{specs})

		err = cmd.Execute()
		assert.NoError(t, err)

		assert.Contains(t, output.String(), meta.Name)
	})

	t.Run("GetValue", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		scrt := &value.Value{
			Name: faker.UUIDHyphenated(),
			Data: faker.UUIDHyphenated(),
		}

		_, err := valueStore.Store(ctx, scrt)
		assert.NoError(t, err)

		output := new(bytes.Buffer)

		cmd := NewGetCommand(GetConfig{
			SpecStore:  specStore,
			ValueStore: valueStore,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{values})

		err = cmd.Execute()
		assert.NoError(t, err)

		assert.Contains(t, output.String(), scrt.Name)
	})
}
