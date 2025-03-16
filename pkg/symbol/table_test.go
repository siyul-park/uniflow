package symbol

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/meta"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/require"
)

func TestTable_AddAndRemoveLoadHook(t *testing.T) {
	tb := NewTable()
	defer tb.Close()

	hook := LoadFunc(func(_ *Symbol) error {
		return nil
	})

	ok := tb.AddLoadHook(hook)
	require.True(t, ok)

	ok = tb.AddLoadHook(hook)
	require.False(t, ok)

	ok = tb.RemoveLoadHook(hook)
	require.True(t, ok)

	ok = tb.RemoveLoadHook(hook)
	require.False(t, ok)
}

func TestTable_AddAndRemoveUnloadHook(t *testing.T) {
	tb := NewTable()
	defer tb.Close()

	hook := UnloadFunc(func(_ *Symbol) error {
		return nil
	})

	ok := tb.AddUnloadHook(hook)
	require.True(t, ok)

	ok = tb.AddUnloadHook(hook)
	require.False(t, ok)

	ok = tb.RemoveUnloadHook(hook)
	require.True(t, ok)

	ok = tb.RemoveUnloadHook(hook)
	require.False(t, ok)
}

func TestTable_Insert(t *testing.T) {
	kind := faker.UUIDHyphenated()

	tb := NewTable()
	defer tb.Close()

	meta1 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: meta.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}
	meta2 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: meta.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}
	meta3 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: meta.DefaultNamespace,
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
	sym2 := &Symbol{Spec: meta2, Node: node.NewOneToOneNode(nil)}
	sym3 := &Symbol{Spec: meta3, Node: node.NewOneToOneNode(nil)}

	err := tb.Insert(sym1)
	require.NoError(t, err)
	require.NotNil(t, tb.Lookup(sym1.ID()))

	err = tb.Insert(sym2)
	require.NoError(t, err)
	require.NotNil(t, tb.Lookup(sym2.ID()))

	err = tb.Insert(sym3)
	require.NoError(t, err)
	require.NotNil(t, tb.Lookup(sym3.ID()))

	p1 := sym1.Out(node.PortOut)
	p2 := sym2.Out(node.PortOut)
	p3 := sym3.Out(node.PortOut)

	require.Len(t, p1.Links(), 1)
	require.Len(t, p2.Links(), 1)
	require.Len(t, p3.Links(), 1)
}

func TestTable_Free(t *testing.T) {
	kind := faker.UUIDHyphenated()

	tb := NewTable()
	defer tb.Close()

	meta1 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: meta.DefaultNamespace,
	}
	meta2 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: meta.DefaultNamespace,
	}
	meta3 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: meta.DefaultNamespace,
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
	sym2 := &Symbol{Spec: meta2, Node: node.NewOneToOneNode(nil)}
	sym3 := &Symbol{Spec: meta3, Node: node.NewOneToOneNode(nil)}

	tb.Insert(sym1)
	tb.Insert(sym2)
	tb.Insert(sym3)

	p1 := sym1.Out(node.PortOut)
	p2 := sym2.Out(node.PortOut)
	p3 := sym3.Out(node.PortOut)

	require.Len(t, p1.Links(), 1)
	require.Len(t, p2.Links(), 1)
	require.Len(t, p3.Links(), 1)

	ok, err := tb.Free(meta1.GetID())
	require.NoError(t, err)
	require.True(t, ok)

	require.Len(t, p1.Links(), 0)

	ok, err = tb.Free(meta2.GetID())
	require.NoError(t, err)
	require.True(t, ok)

	require.Len(t, p1.Links(), 0)
	require.Len(t, p2.Links(), 0)

	ok, err = tb.Free(meta3.GetID())
	require.NoError(t, err)
	require.True(t, ok)

	require.Len(t, p1.Links(), 0)
	require.Len(t, p2.Links(), 0)
	require.Len(t, p3.Links(), 0)
}

func TestTable_Lookup(t *testing.T) {
	kind := faker.UUIDHyphenated()

	tb := NewTable()
	defer tb.Close()

	meta := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: meta.DefaultNamespace,
	}

	sb := &Symbol{Spec: meta, Node: node.NewOneToOneNode(nil)}

	tb.Insert(sb)
	require.Equal(t, sb, tb.Lookup(sb.ID()))
}

func TestTable_Keys(t *testing.T) {
	kind := faker.UUIDHyphenated()

	tb := NewTable()
	defer tb.Close()

	meta := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: meta.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}
	sb := &Symbol{Spec: meta, Node: node.NewOneToOneNode(nil)}

	tb.Insert(sb)

	ids := tb.Keys()
	require.Contains(t, ids, sb.ID())
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
		Namespace: meta.DefaultNamespace,
	}
	meta2 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: meta.DefaultNamespace,
	}
	meta3 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: meta.DefaultNamespace,
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
	sym2 := &Symbol{Spec: meta2, Node: node.NewOneToOneNode(nil)}
	sym3 := &Symbol{Spec: meta3, Node: node.NewOneToOneNode(nil)}

	err := tb.Insert(sym1)
	require.NoError(t, err)
	require.Equal(t, 0, loaded)
	require.Equal(t, 0, unloaded)

	err = tb.Insert(sym2)
	require.NoError(t, err)
	require.Equal(t, 0, loaded)
	require.Equal(t, 0, unloaded)

	err = tb.Insert(sym3)
	require.NoError(t, err)
	require.Equal(t, 3, loaded)
	require.Equal(t, 0, unloaded)

	_, err = tb.Free(sym1.ID())
	require.NoError(t, err)
	require.Equal(t, 3, loaded)
	require.Equal(t, 3, unloaded)

	_, err = tb.Free(sym2.ID())
	require.NoError(t, err)
	require.Equal(t, 3, loaded)
	require.Equal(t, 3, unloaded)

	_, err = tb.Free(sym3.ID())
	require.NoError(t, err)
	require.Equal(t, 3, loaded)
	require.Equal(t, 3, unloaded)
}

func BenchmarkTable_Insert(b *testing.B) {
	kind := faker.UUIDHyphenated()

	tb := NewTable()
	defer tb.Close()

	for i := 0; i < b.N; i++ {
		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: meta.DefaultNamespace,
		}

		sb := &Symbol{Spec: meta, Node: node.NewOneToOneNode(nil)}
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
			Namespace: meta.DefaultNamespace,
		}

		sb := &Symbol{Spec: meta, Node: node.NewOneToOneNode(nil)}
		_ = tb.Insert(sb)

		b.StartTimer()

		_, _ = tb.Free(meta.GetID())
	}
}
