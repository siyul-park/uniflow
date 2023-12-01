package resource

import (
	"encoding/json"
	"testing"
	"testing/fstest"

	"github.com/go-faker/faker/v4"
	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/assert"
)

func TestBuilder_Build(t *testing.T) {
	s := scheme.New()
	fsys := make(fstest.MapFS)

	filename := "spec.json"
	kind := faker.Word()

	spec := &scheme.SpecMeta{
		ID:        ulid.Make(),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
	}

	codec := scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(node.OneToOneNodeConfig{ID: spec.GetID()}), nil
	})

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, codec)

	data, _ := json.Marshal(spec)

	fsys[filename] = &fstest.MapFile{
		Data: data,
	}

	builder := NewBuilder().
		Scheme(s).
		Namespace(scheme.DefaultNamespace).
		FS(fsys).
		Filename(filename)

	specs, err := builder.Build()
	assert.NoError(t, err)
	assert.Len(t, specs, 1)
}
