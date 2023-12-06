package start

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"testing/fstest"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	s := scheme.New()
	h := hook.New()
	db := memdb.New("")
	fsys := make(fstest.MapFS)

	st, _ := storage.New(context.Background(), storage.Config{
		Scheme:   s,
		Database: db,
	})

	bootFilepath := "boot.json"
	kind := faker.Word()

	spec := &scheme.SpecMeta{
		ID:        ulid.Make(),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
	}

	codec := scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, codec)

	data, _ := json.Marshal(spec)

	fsys[bootFilepath] = &fstest.MapFile{
		Data: data,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	output := new(bytes.Buffer)

	cmd := NewCmd(Config{
		Scheme:   s,
		Hook:     h,
		FS:       fsys,
		Database: db,
	})
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetContext(ctx)

	cmd.SetArgs([]string{"--boot", bootFilepath})

	go func() {
		_ = cmd.Execute()
	}()

	for {
		select {
		case <-ctx.Done():
			assert.Fail(t, "timeout")
			return
		default:
			r, err := st.FindOne(ctx, storage.Where[ulid.ULID](scheme.KeyID).EQ(spec.GetID()))
			assert.NoError(t, err)
			if r != nil {
				return
			}

			// TODO: assert symbol is loaded.
		}
	}
}
