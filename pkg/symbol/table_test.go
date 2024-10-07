package symbol

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestTable_Insert(t *testing.T) {
	kind := faker.UUIDHyphenated()

	tb := NewTable()
	defer tb.Close()

	meta1 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}
	meta2 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}
	meta3 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	meta1.Ports = map[string][]spec.Port{
		node.PortOut: {
			{
				ID:   meta2.GetID(),
				Port: node.PortIn,
			},
		},
	}
	meta2.Ports = map[string][]spec.Port{
		node.PortOut: {
			{
				Name: meta3.GetName(),
				Port: node.PortIn,
			},
		},
	}
	meta3.Ports = map[string][]spec.Port{
		node.PortOut: {
			{
				ID:   meta1.GetID(),
				Port: node.PortIn,
			},
		},
	}

	sym1 := &Symbol{Spec: meta1, Node: node.NewOneToOneNode(nil)}
	err := tb.Insert(sym1)
	assert.NoError(t, err)
	assert.NotNil(t, tb.Lookup(sym1.ID()))

	sym2 := &Symbol{Spec: meta2, Node: node.NewOneToOneNode(nil)}
	err = tb.Insert(sym2)
	assert.NoError(t, err)
	assert.NotNil(t, tb.Lookup(sym2.ID()))

	sym3 := &Symbol{Spec: meta3, Node: node.NewOneToOneNode(nil)}
	err = tb.Insert(sym3)
	assert.NoError(t, err)
	assert.NotNil(t, tb.Lookup(sym3.ID()))

	p1 := sym1.Out(node.PortOut)
	p2 := sym2.Out(node.PortOut)
	p3 := sym3.Out(node.PortOut)

	assert.Equal(t, 1, p1.Links())
	assert.Equal(t, 1, p2.Links())
	assert.Equal(t, 1, p3.Links())
}

func TestTable_Free(t *testing.T) {
	kind := faker.UUIDHyphenated()

	tb := NewTable()
	defer tb.Close()

	meta1 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: resource.DefaultNamespace,
	}
	meta2 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: resource.DefaultNamespace,
	}
	meta3 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: resource.DefaultNamespace,
	}

	meta1.Ports = map[string][]spec.Port{
		node.PortOut: {
			{
				ID:   meta2.GetID(),
				Port: node.PortIn,
			},
		},
	}
	meta2.Ports = map[string][]spec.Port{
		node.PortOut: {
			{
				ID:   meta3.GetID(),
				Port: node.PortIn,
			},
		},
	}
	meta3.Ports = map[string][]spec.Port{
		node.PortOut: {
			{
				ID:   meta1.GetID(),
				Port: node.PortIn,
			},
		},
	}

	sym1 := &Symbol{Spec: meta1, Node: node.NewOneToOneNode(nil)}
	err := tb.Insert(sym1)
	assert.NoError(t, err)

	sym2 := &Symbol{Spec: meta2, Node: node.NewOneToOneNode(nil)}
	err = tb.Insert(sym2)
	assert.NoError(t, err)

	sym3 := &Symbol{Spec: meta3, Node: node.NewOneToOneNode(nil)}
	err = tb.Insert(sym3)
	assert.NoError(t, err)

	p1 := sym1.Out(node.PortOut)
	p2 := sym2.Out(node.PortOut)
	p3 := sym3.Out(node.PortOut)

	assert.Equal(t, 1, p1.Links())
	assert.Equal(t, 1, p2.Links())
	assert.Equal(t, 1, p3.Links())

	ok, err := tb.Free(meta1.GetID())
	assert.NoError(t, err)
	assert.True(t, ok)

	assert.Equal(t, 0, p1.Links())

	ok, err = tb.Free(meta2.GetID())
	assert.NoError(t, err)
	assert.True(t, ok)

	assert.Equal(t, 0, p1.Links())
	assert.Equal(t, 0, p2.Links())

	ok, err = tb.Free(meta3.GetID())
	assert.NoError(t, err)
	assert.True(t, ok)

	assert.Equal(t, 0, p1.Links())
	assert.Equal(t, 0, p2.Links())
	assert.Equal(t, 0, p3.Links())
}

func TestTable_Lookup(t *testing.T) {
	kind := faker.UUIDHyphenated()

	tb := NewTable()
	defer tb.Close()

	meta := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: resource.DefaultNamespace,
	}

	sb := &Symbol{Spec: meta, Node: node.NewOneToOneNode(nil)}

	err := tb.Insert(sb)
	assert.NoError(t, err)
	assert.Equal(t, sb, tb.Lookup(sb.ID()))
}

func TestTable_Keys(t *testing.T) {
	kind := faker.UUIDHyphenated()

	tb := NewTable()
	defer tb.Close()

	meta := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	sb := &Symbol{Spec: meta, Node: node.NewOneToOneNode(nil)}
	err := tb.Insert(sb)
	assert.NoError(t, err)

	ids := tb.Keys()
	assert.Contains(t, ids, sb.ID())
}

func TestTable_Hook(t *testing.T) {
	kind := faker.UUIDHyphenated()

	loaded := 0
	unloaded := 0

	tb := NewTable(TableOption{
		LoadHooks: []LoadHook{
			LoadFunc(func(_ *Symbol) error {
				loaded += 1
				return nil
			}),
		},
		UnloadHooks: []UnloadHook{
			UnloadFunc(func(_ *Symbol) error {
				unloaded += 1
				return nil
			}),
		},
	})
	defer tb.Close()

	meta1 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: resource.DefaultNamespace,
	}
	meta2 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: resource.DefaultNamespace,
	}
	meta3 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: resource.DefaultNamespace,
	}

	meta1.Ports = map[string][]spec.Port{
		node.PortOut: {
			{
				ID:   meta2.GetID(),
				Port: node.PortIn,
			},
		},
	}
	meta2.Ports = map[string][]spec.Port{
		node.PortOut: {
			{
				ID:   meta3.GetID(),
				Port: node.PortIn,
			},
		},
	}
	meta3.Ports = map[string][]spec.Port{
		node.PortOut: {
			{
				ID:   meta1.GetID(),
				Port: node.PortIn,
			},
		},
	}
	sym1 := &Symbol{Spec: meta1, Node: node.NewOneToOneNode(nil)}
	err := tb.Insert(sym1)
	assert.NoError(t, err)
	assert.Equal(t, 0, loaded)
	assert.Equal(t, 0, unloaded)

	sym2 := &Symbol{Spec: meta2, Node: node.NewOneToOneNode(nil)}
	err = tb.Insert(sym2)
	assert.NoError(t, err)
	assert.Equal(t, 0, loaded)
	assert.Equal(t, 0, unloaded)

	sym3 := &Symbol{Spec: meta3, Node: node.NewOneToOneNode(nil)}
	err = tb.Insert(sym3)
	assert.NoError(t, err)
	assert.Equal(t, 3, loaded)
	assert.Equal(t, 0, unloaded)

	_, err = tb.Free(sym1.ID())
	assert.NoError(t, err)
	assert.Equal(t, 3, loaded)
	assert.Equal(t, 3, unloaded)

	_, err = tb.Free(sym2.ID())
	assert.NoError(t, err)
	assert.Equal(t, 3, loaded)
	assert.Equal(t, 3, unloaded)

	_, err = tb.Free(sym3.ID())
	assert.NoError(t, err)
	assert.Equal(t, 3, loaded)
	assert.Equal(t, 3, unloaded)
}

func BenchmarkTable_Insert(b *testing.B) {
	kind := faker.UUIDHyphenated()

	tb := NewTable()
	defer tb.Close()

	for i := 0; i < b.N; i++ {
		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
		}

		sb := &Symbol{Spec: meta}
		_ = tb.Insert(sb)
	}
}

func BenchmarkTable_Free(b *testing.B) {
	kind := faker.UUIDHyphenated()

	tb := NewTable()
	defer tb.Close()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
		}

		sb := &Symbol{Spec: meta}
		_ = tb.Insert(sb)

		b.StartTimer()

		_, _ = tb.Free(meta.GetID())
	}
}
