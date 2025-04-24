package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/store"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/hook"
	"github.com/siyul-park/uniflow/meta"
	"github.com/siyul-park/uniflow/node"
	"github.com/siyul-park/uniflow/scheme"
	"github.com/siyul-park/uniflow/spec"
	testingutil "github.com/siyul-park/uniflow/testing"
	"github.com/siyul-park/uniflow/value"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestTestCommand_Execute(t *testing.T) {
	r := testingutil.NewRunner()

	s := scheme.New()
	h := hook.New()

	specStore := store.New()
	valueStore := store.New()

	fs := afero.NewMemMapFs()

	kind := faker.UUIDHyphenated()

	codec := scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, codec)

	t.Run("NoFlag", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: meta.DefaultNamespace,
		}

		err := specStore.Insert(ctx, []any{meta})
		require.NoError(t, err)

		h := hook.New()

		output := new(bytes.Buffer)

		cmd := NewTestCommand(TestConfig{
			Runner:     r,
			Scheme:     s,
			Hook:       h,
			FS:         fs,
			SpecStore:  specStore,
			ValueStore: valueStore,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetContext(ctx)

		err = cmd.Execute()
		require.NoError(t, err)
	})

	t.Run("Regexp", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: meta.DefaultNamespace,
		}

		err := specStore.Insert(ctx, []any{meta})
		require.NoError(t, err)

		h := hook.New()

		output := new(bytes.Buffer)

		cmd := NewTestCommand(TestConfig{
			Runner:     r,
			Scheme:     s,
			Hook:       h,
			FS:         fs,
			SpecStore:  specStore,
			ValueStore: valueStore,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetContext(ctx)

		cmd.SetArgs([]string{"foo"})

		err = cmd.Execute()
		require.NoError(t, err)
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

		cmd := NewTestCommand(TestConfig{
			Runner:     r,
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

		cmd := NewTestCommand(TestConfig{
			Runner:     r,
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
