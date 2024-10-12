package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/chart"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestStartCommand_Execute(t *testing.T) {
	s := scheme.New()
	h := hook.New()

	specStore := spec.NewStore()
	secretStore := secret.NewStore()
	chartStore := chart.NewStore()

	fs := afero.NewMemMapFs()

	kind := faker.UUIDHyphenated()

	codec := scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, codec)

	t.Run("NoFlag", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
		}

		specStore.Store(ctx, meta)

		h := hook.New()
		symbols := make(chan *symbol.Symbol)

		h.AddLoadHook(symbol.LoadFunc(func(sb *symbol.Symbol) error {
			symbols <- sb
			return nil
		}))

		output := new(bytes.Buffer)

		cmd := NewStartCommand(StartConfig{
			Scheme:      s,
			Hook:        h,
			FS:          fs,
			SpecStore:   specStore,
			SecretStore: secretStore,
			ChartStore:  chartStore,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetContext(ctx)

		go func() {
			_ = cmd.Execute()
		}()

		select {
		case <-symbols:
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run(flagDebug, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
		}

		specStore.Store(ctx, meta)

		h := hook.New()
		symbols := make(chan *symbol.Symbol)

		h.AddLoadHook(symbol.LoadFunc(func(sb *symbol.Symbol) error {
			symbols <- sb
			return nil
		}))

		output := new(bytes.Buffer)

		cmd := NewStartCommand(StartConfig{
			Scheme:      s,
			Hook:        h,
			FS:          fs,
			SpecStore:   specStore,
			SecretStore: secretStore,
			ChartStore:  chartStore,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetContext(ctx)

		cmd.SetArgs([]string{fmt.Sprintf("--%s", flagDebug)})

		go func() {
			_ = cmd.Execute()
		}()

		select {
		case <-symbols:
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run(flagFromSpecs, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		filename := "specs.json"

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
		}

		data, _ := json.Marshal(meta)

		f, _ := fs.Create(filename)
		f.Write(data)

		output := new(bytes.Buffer)

		cmd := NewStartCommand(StartConfig{
			Scheme:      s,
			Hook:        h,
			FS:          fs,
			SpecStore:   specStore,
			SecretStore: secretStore,
			ChartStore:  chartStore,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetContext(ctx)

		cmd.SetArgs([]string{fmt.Sprintf("--%s", flagFromSpecs), filename})

		specStream, _ := specStore.Watch(ctx)
		defer specStream.Close()

		go func() {
			_ = cmd.Execute()
		}()

		select {
		case <-specStream.Next():
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run(flagFromSecrets, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		filename := "secrets.json"

		scrt := &secret.Secret{
			ID:        uuid.Must(uuid.NewV7()),
			Namespace: resource.DefaultNamespace,
			Data:      faker.Word(),
		}

		data, _ := json.Marshal(scrt)

		f, _ := fs.Create(filename)
		f.Write(data)

		output := new(bytes.Buffer)

		cmd := NewStartCommand(StartConfig{
			Scheme:      s,
			Hook:        h,
			FS:          fs,
			SpecStore:   specStore,
			SecretStore: secretStore,
			ChartStore:  chartStore,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetContext(ctx)

		cmd.SetArgs([]string{fmt.Sprintf("--%s", flagFromSecrets), filename})

		secretStream, _ := secretStore.Watch(ctx)
		defer secretStream.Close()

		go func() {
			_ = cmd.Execute()
		}()

		select {
		case <-secretStream.Next():
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run(flagFromCharts, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		filename := "charts.json"

		chrt := &chart.Chart{
			ID:        uuid.Must(uuid.NewV7()),
			Namespace: resource.DefaultNamespace,
			Name:      faker.Word(),
		}

		data, _ := json.Marshal(chrt)

		f, _ := fs.Create(filename)
		f.Write(data)

		output := new(bytes.Buffer)

		cmd := NewStartCommand(StartConfig{
			Scheme:      s,
			Hook:        h,
			FS:          fs,
			SpecStore:   specStore,
			SecretStore: secretStore,
			ChartStore:  chartStore,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetContext(ctx)

		cmd.SetArgs([]string{fmt.Sprintf("--%s", flagFromCharts), filename})

		chartStream, _ := chartStore.Watch(ctx)
		defer chartStream.Close()

		go func() {
			_ = cmd.Execute()
		}()

		select {
		case <-chartStream.Next():
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})
}
