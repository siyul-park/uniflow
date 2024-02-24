package scanner

import (
	"context"
	"encoding/json"
	"testing"
	"testing/fstest"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestScanner_Scan(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := scheme.New()
	db := memdb.New("")
	fsys := make(fstest.MapFS)

	kind := faker.UUIDHyphenated()

	st, _ := storage.New(ctx, storage.Config{
		Scheme:   s,
		Database: db,
	})

	codec := scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, codec)

	filename := "spec.json"

	spec := &scheme.SpecMeta{
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	data, _ := json.Marshal(spec)

	_, _ = st.InsertOne(ctx, spec)

	fsys[filename] = &fstest.MapFile{
		Data: data,
	}

	scanner := New().
		Scheme(s).
		Storage(st).
		Namespace(scheme.DefaultNamespace).
		FS(fsys).
		Filename(filename)

	specs, err := scanner.Scan(ctx)
	assert.NoError(t, err)
	assert.Len(t, specs, 1)
}
