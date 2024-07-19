package scanner

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestScanner_Scan(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := scheme.New()
	st := spec.NewStore(memdb.NewCollection(""))
	fsys := afero.NewMemMapFs()

	kind := faker.UUIDHyphenated()

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

	_, _ = st.Store(ctx, meta)

	f, _ := fsys.Create(filename)
	f.Write(data)

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
