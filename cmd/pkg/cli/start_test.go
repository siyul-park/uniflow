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
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestStartCommand_Execute(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := scheme.New()
	h := hook.New()
	spst := spec.NewStore()
	scst := secret.NewStore()
	fsys := afero.NewMemMapFs()

	kind := faker.UUIDHyphenated()

	codec := scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, codec)

	filename := "resources.json"

	meta := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: spec.DefaultNamespace,
	}

	data, _ := json.Marshal(meta)

	f, _ := fsys.Create(filename)
	f.Write(data)

	func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		output := new(bytes.Buffer)

		cmd := NewStartCommand(StartConfig{
			Scheme:      s,
			Hook:        h,
			FS:          fsys,
			SpecStore:   spst,
			SecretStore: scst,
		})
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetContext(ctx)

		cmd.SetArgs([]string{fmt.Sprintf("--%s", flagFilename), filename})

		go func() {
			_ = cmd.Execute()
		}()

		for {
			select {
			case <-ctx.Done():
				assert.Fail(t, "timeout")
				return
			default:
				if r, _ := spst.Load(ctx, meta); len(r) > 0 {
					return
				}
			}
		}
	}()
}
