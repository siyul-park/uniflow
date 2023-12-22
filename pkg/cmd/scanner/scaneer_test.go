package scanner

import (
	"context"
	"encoding/json"
	"testing"
	"testing/fstest"

	"github.com/go-faker/faker/v4"
	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestScanner_Scan(t *testing.T) {
	s := scheme.New()
	db := memdb.New("")
	fsys := make(fstest.MapFS)

	st, _ := storage.New(context.Background(), storage.Config{
		Scheme:   s,
		Database: db,
	})

	filename := "spec.json"
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

	fsys[filename] = &fstest.MapFile{
		Data: data,
	}

	scanner := New().
		Scheme(s).
		Storage(st).
		Namespace(scheme.DefaultNamespace).
		FS(fsys).
		Filename(filename)

	specs, err := scanner.Scan(context.Background())
	assert.NoError(t, err)
	assert.Len(t, specs, 1)
}
