package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/pkg/driver"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/meta"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/siyul-park/uniflow/pkg/value"
)

func TestStartCommand_Execute(t *testing.T) {
	s := scheme.New()
	h := hook.New()

	specStore := driver.NewStore()
	valueStore := driver.NewStore()

	fs := afero.NewMemMapFs()

	kind := faker.UUIDHyphenated()

	codec := scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, codec)

	t.Run("NoFlag", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
		defer cancel()

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: meta.DefaultNamespace,
		}

		err := specStore.Insert(ctx, []any{meta})
		require.NoError(t, err)

		h := hook.New()
		symbols := make(chan *symbol.Symbol)

		h.AddLoadHook(symbol.LoadFunc(func(sb *symbol.Symbol) error {
			symbols <- sb
			return nil
		}))

		output := new(bytes.Buffer)

		cmd := NewStartCommand(StartConfig{
			Scheme:     s,
			Hook:       h,
			FS:         fs,
			SpecStore:  specStore,
			ValueStore: valueStore,
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
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run(flagDebug, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
		defer cancel()

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: meta.DefaultNamespace,
		}

		err := specStore.Insert(ctx, []any{meta})
		require.NoError(t, err)

		h := hook.New()
		symbols := make(chan *symbol.Symbol)

		h.AddLoadHook(symbol.LoadFunc(func(sb *symbol.Symbol) error {
			symbols <- sb
			return nil
		}))

		output := new(bytes.Buffer)

		cmd := NewStartCommand(StartConfig{
			Scheme:     s,
			Hook:       h,
			FS:         fs,
			SpecStore:  specStore,
			ValueStore: valueStore,
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
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run(flagFromSpecs, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		filename := "specs.json"

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: meta.DefaultNamespace,
		}

		data, _ := json.Marshal(meta)

		f, _ := fs.Create(filename)
		f.Write(data)

		output := new(bytes.Buffer)

		cmd := NewStartCommand(StartConfig{
			Scheme:     s,
			Hook:       h,
			FS:         fs,
			SpecStore:  specStore,
			ValueStore: valueStore,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetContext(ctx)

		cmd.SetArgs([]string{fmt.Sprintf("--%s", flagFromSpecs), filename})

		strm, err := specStore.Watch(ctx, nil)
		require.NoError(t, err)
		require.NotNil(t, strm)

		defer strm.Close(ctx)

		var count atomic.Int32
		go func() {
			for strm.Next(ctx) {
				count.Add(1)
			}
		}()

		go func() {
			_ = cmd.Execute()
		}()

		require.Eventually(t, func() bool { return count.Load() == 1 }, time.Second, 10*time.Millisecond)
	})

	t.Run(flagFromValues, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		filename := "values.json"

		scrt := &value.Value{
			ID:        uuid.Must(uuid.NewV7()),
			Namespace: meta.DefaultNamespace,
			Data:      faker.UUIDHyphenated(),
		}

		data, _ := json.Marshal(scrt)

		f, _ := fs.Create(filename)
		f.Write(data)

		output := new(bytes.Buffer)

		cmd := NewStartCommand(StartConfig{
			Scheme:     s,
			Hook:       h,
			FS:         fs,
			SpecStore:  specStore,
			ValueStore: valueStore,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetContext(ctx)

		cmd.SetArgs([]string{fmt.Sprintf("--%s", flagFromValues), filename})

		strm, err := valueStore.Watch(ctx, nil)
		require.NoError(t, err)
		require.NotNil(t, strm)

		defer strm.Close(ctx)

		var count atomic.Int32
		go func() {
			for strm.Next(ctx) {
				count.Add(1)
			}
		}()

		go func() {
			_ = cmd.Execute()
		}()

		require.Eventually(t, func() bool { return count.Load() == 1 }, time.Second, 10*time.Millisecond)
	})
}
