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
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	testingutil "github.com/siyul-park/uniflow/pkg/testing"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestTestCommand_Execute(t *testing.T) {
	r := testingutil.NewRunner(nil)

	s := scheme.New()
	h := hook.New()

	specStore := spec.NewStore()
	valueStore := value.NewStore()

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
			Namespace: resource.DefaultNamespace,
		}

		specStore.Store(ctx, meta)

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

		err := cmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("Regexp", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
		}

		specStore.Store(ctx, meta)

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

		err := cmd.Execute()
		assert.NoError(t, err)
	})

	t.Run(flagFromSpecs, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
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

		specStream, _ := specStore.Watch(ctx)
		defer specStream.Close()

		err := cmd.Execute()
		assert.NoError(t, err)

		select {
		case <-specStream.Next():
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run(flagFromValues, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		filename := "values.json"

		scrt := &value.Value{
			ID:        uuid.Must(uuid.NewV7()),
			Namespace: resource.DefaultNamespace,
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

		valueStream, _ := valueStore.Watch(ctx)
		defer valueStream.Close()

		err := cmd.Execute()
		assert.NoError(t, err)

		select {
		case <-valueStream.Next():
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})
}
