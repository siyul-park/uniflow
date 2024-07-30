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
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestStartCommand_Execute(t *testing.T) {
	s := scheme.New()
	h := hook.New()

	specStore := spec.NewStore()
	secretStore := secret.NewStore()

	fsys := afero.NewMemMapFs()

	kind := faker.UUIDHyphenated()

	codec := scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, codec)

	t.Run("ExecuteFromNodes", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		filename := "nodes.json"

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
		}

		data, _ := json.Marshal(meta)

		f, _ := fsys.Create(filename)
		f.Write(data)

		output := new(bytes.Buffer)

		cmd := NewStartCommand(StartConfig{
			Scheme:      s,
			Hook:        h,
			FS:          fsys,
			SpecStore:   specStore,
			SecretStore: secretStore,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetContext(ctx)

		cmd.SetArgs([]string{fmt.Sprintf("--%s", flagFromNodes), filename})

		go func() {
			_ = cmd.Execute()
		}()

		for {
			select {
			case <-ctx.Done():
				assert.Fail(t, ctx.Err().Error())
				return
			default:
				if r, _ := specStore.Load(ctx, meta); len(r) > 0 {
					return
				}
			}
		}
	})

	t.Run("ExecuteFromSecrets", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		filename := "nodes.json"

		sec := &secret.Secret{
			ID:        uuid.Must(uuid.NewV7()),
			Namespace: resource.DefaultNamespace,
			Data:      faker.Word(),
		}

		data, _ := json.Marshal(sec)

		f, _ := fsys.Create(filename)
		f.Write(data)

		output := new(bytes.Buffer)

		cmd := NewStartCommand(StartConfig{
			Scheme:      s,
			Hook:        h,
			FS:          fsys,
			SpecStore:   specStore,
			SecretStore: secretStore,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetContext(ctx)

		cmd.SetArgs([]string{fmt.Sprintf("--%s", flagFromSecrets), filename})

		go func() {
			_ = cmd.Execute()
		}()

		select {
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
			return
		default:
			if r, _ := secretStore.Load(ctx, sec); len(r) > 0 {
				return
			}
		}
	})
}
