package scanner

import (
	"context"
	"encoding/json"
	"testing"
	"testing/fstest"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestScanner_Scan(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := spec.NewScheme()
	db := memdb.New("")
	fsys := make(fstest.MapFS)

	kind := faker.UUIDHyphenated()

	st, _ := spec.NewStorage(ctx, spec.StorageConfig{
		Scheme:   s,
		Database: db,
	})

	codec := spec.CodecFunc(func(spec spec.Spec) (node.Node, error) {
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
		Storage(st).
		Namespace(spec.DefaultNamespace).
		FS(fsys).
		Filename(filename)

	specs, err := scanner.Scan(ctx)
	assert.NoError(t, err)
	assert.Len(t, specs, 1)
}
