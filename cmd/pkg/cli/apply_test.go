package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/chart"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestApplyCommand_Execute(t *testing.T) {
	chartStore := chart.NewStore()
	specStore := spec.NewStore()
	secretStore := secret.NewStore()

	fs := afero.NewMemMapFs()

	t.Run("InsertChart", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "chart.json"

		chrt := &chart.Chart{
			ID:        uuid.Must(uuid.NewV7()),
			Namespace: resource.DefaultNamespace,
			Name:      faker.Word(),
		}

		data, err := json.Marshal(chrt)
		assert.NoError(t, err)

		file, err := fs.Create(filename)
		assert.NoError(t, err)
		defer file.Close()

		_, err = file.Write(data)
		assert.NoError(t, err)

		output := new(bytes.Buffer)

		cmd := NewApplyCommand(ApplyConfig{
			ChartStore:  chartStore,
			SpecStore:   specStore,
			SecretStore: secretStore,
			FS:          fs,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{argCharts, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		assert.NoError(t, err)

		results, err := chartStore.Load(ctx, chrt)
		assert.NoError(t, err)
		assert.Len(t, results, 1)

		assert.Contains(t, output.String(), chrt.Name)
	})

	t.Run("UpdateChart", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "chart.json"

		chrt := &chart.Chart{
			ID:        uuid.Must(uuid.NewV7()),
			Namespace: resource.DefaultNamespace,
			Name:      faker.Word(),
		}

		_, err := chartStore.Store(ctx, chrt)
		assert.NoError(t, err)

		data, err := json.Marshal(chrt)
		assert.NoError(t, err)

		file, err := fs.Create(filename)
		assert.NoError(t, err)
		defer file.Close()

		_, err = file.Write(data)
		assert.NoError(t, err)

		output := new(bytes.Buffer)

		cmd := NewApplyCommand(ApplyConfig{
			ChartStore:  chartStore,
			SpecStore:   specStore,
			SecretStore: secretStore,
			FS:          fs,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{argCharts, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		assert.NoError(t, err)

		results, err := chartStore.Load(ctx, chrt)
		assert.NoError(t, err)
		assert.Len(t, results, 1)

		assert.Contains(t, output.String(), chrt.Name)
	})

	t.Run("InsertNodeSpec", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "nodes.json"

		kind := faker.UUIDHyphenated()

		meta := &spec.Meta{
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		}

		data, err := json.Marshal(meta)
		assert.NoError(t, err)

		file, err := fs.Create(filename)
		assert.NoError(t, err)
		defer file.Close()

		_, err = file.Write(data)
		assert.NoError(t, err)

		output := new(bytes.Buffer)

		cmd := NewApplyCommand(ApplyConfig{
			ChartStore:  chartStore,
			SpecStore:   specStore,
			SecretStore: secretStore,
			FS:          fs,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{argNodes, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		assert.NoError(t, err)

		results, err := specStore.Load(ctx, meta)
		assert.NoError(t, err)
		assert.Len(t, results, 1)

		assert.Contains(t, output.String(), meta.Name)
	})

	t.Run("UpdateNodeSpec", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "nodes.json"

		kind := faker.UUIDHyphenated()

		meta := &spec.Meta{
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		}

		_, err := specStore.Store(ctx, meta)
		assert.NoError(t, err)

		data, err := json.Marshal(meta)
		assert.NoError(t, err)

		file, err := fs.Create(filename)
		assert.NoError(t, err)
		defer file.Close()

		_, err = file.Write(data)
		assert.NoError(t, err)

		output := new(bytes.Buffer)

		cmd := NewApplyCommand(ApplyConfig{
			ChartStore:  chartStore,
			SpecStore:   specStore,
			SecretStore: secretStore,
			FS:          fs,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{argNodes, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		assert.NoError(t, err)

		results, err := specStore.Load(ctx, meta)
		assert.NoError(t, err)
		assert.Len(t, results, 1)

		assert.Contains(t, output.String(), meta.Name)
	})

	t.Run("InsertSecret", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "secrets.json"

		scrt := &secret.Secret{
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Data:      faker.Word(),
		}

		data, err := json.Marshal(scrt)
		assert.NoError(t, err)

		file, err := fs.Create(filename)
		assert.NoError(t, err)
		defer file.Close()

		_, err = file.Write(data)
		assert.NoError(t, err)

		output := new(bytes.Buffer)

		cmd := NewApplyCommand(ApplyConfig{
			ChartStore:  chartStore,
			SpecStore:   specStore,
			SecretStore: secretStore,
			FS:          fs,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{argSecrets, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		assert.NoError(t, err)

		results, err := secretStore.Load(ctx, scrt)
		assert.NoError(t, err)
		assert.Len(t, results, 1)

		assert.Contains(t, output.String(), scrt.Name)
	})

	t.Run("UpdateSecret", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		filename := "secrets.json"

		scrt := &secret.Secret{
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
			Data:      faker.Word(),
		}

		_, err := secretStore.Store(ctx, scrt)
		assert.NoError(t, err)

		data, err := json.Marshal(scrt)
		assert.NoError(t, err)

		file, err := fs.Create(filename)
		assert.NoError(t, err)
		defer file.Close()

		_, err = file.Write(data)
		assert.NoError(t, err)

		output := new(bytes.Buffer)

		cmd := NewApplyCommand(ApplyConfig{
			ChartStore:  chartStore,
			SpecStore:   specStore,
			SecretStore: secretStore,
			FS:          fs,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{argSecrets, fmt.Sprintf("--%s", flagFilename), filename})

		err = cmd.Execute()
		assert.NoError(t, err)

		results, err := secretStore.Load(ctx, scrt)
		assert.NoError(t, err)
		assert.Len(t, results, 1)

		assert.Contains(t, output.String(), scrt.Name)
	})
}
