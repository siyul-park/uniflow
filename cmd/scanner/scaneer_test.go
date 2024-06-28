package scanner

import (
	"context"
	"encoding/json"
	"testing"
	"testing/fstest"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/database/memdb"
	"github.com/siyul-park/uniflow/node"
	"github.com/siyul-park/uniflow/scheme"
	"github.com/siyul-park/uniflow/spec"
	"github.com/siyul-park/uniflow/store"
	"github.com/stretchr/testify/assert"
)

func TestScanner_Scan(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := scheme.New()
	db := memdb.New("")
	fsys := make(fstest.MapFS)

	kind := faker.UUIDHyphenated()

	st, _ := store.New(ctx, store.Config{
		Scheme:   s,
		Database: db,
	})

	codec := scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, codec)

	filename := "spec.json"

	meta := &spec.Meta{
		Kind:      kind,
		Namespace: spec.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	data, _ := json.Marshal(meta)

	_, _ = st.InsertOne(ctx, meta)

	fsys[filename] = &fstest.MapFile{
		Data: data,
	}

	scanner := New().
		Scheme(s).
		Store(st).
		Namespace(spec.DefaultNamespace).
		FS(fsys).
		Filename(filename)

	specs, err := scanner.Scan(ctx)
	assert.NoError(t, err)
	assert.Len(t, specs, 1)
}
