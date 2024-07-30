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

	t.Run("Link By ID", func(t *testing.T) {
		t.Run("Unlinked", func(t *testing.T) {
			tb := NewTable()
			defer tb.Clear()

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
		})

		t.Run("Linked", func(t *testing.T) {
			tb := NewTable()
			defer tb.Clear()

			id := uuid.Must(uuid.NewV7())

			meta1 := &spec.Meta{
				ID:        id,
				Kind:      kind,
				Namespace: resource.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}
			meta2 := &spec.Meta{
				ID:        id,
				Kind:      kind,
				Namespace: resource.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}
			meta3 := &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: resource.DefaultNamespace,
			}
			meta4 := &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: resource.DefaultNamespace,
			}

			meta1.Ports = map[string][]spec.Port{
				node.PortOut: {
					{
						ID:   meta3.GetID(),
						Port: node.PortIn,
					},
				},
			}
			meta2.Ports = map[string][]spec.Port{
				node.PortOut: {
					{
						ID:   meta4.GetID(),
						Port: node.PortIn,
					},
				},
			}

			sym3 := &Symbol{Spec: meta3, Node: node.NewOneToOneNode(nil)}
			err := tb.Insert(sym3)
			assert.NoError(t, err)

			sym4 := &Symbol{Spec: meta4, Node: node.NewOneToOneNode(nil)}
			err = tb.Insert(sym4)
			assert.NoError(t, err)

			sym1 := &Symbol{Spec: meta1, Node: node.NewOneToOneNode(nil)}
			err = tb.Insert(sym1)
			assert.NoError(t, err)

			p1 := sym1.Out(node.PortOut)
			assert.Equal(t, 1, p1.Links())

			sym2 := &Symbol{Spec: meta2, Node: node.NewOneToOneNode(nil)}
			err = tb.Insert(sym2)
			assert.NoError(t, err)

			p2 := sym2.Out(node.PortOut)
			assert.Equal(t, 1, p2.Links())
		})
	})

	t.Run("Link By Name", func(t *testing.T) {
		t.Run("Unlinked", func(t *testing.T) {
			tb := NewTable()
			defer tb.Clear()

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
						Name: meta2.GetName(),
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
						Name: meta1.GetName(),
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
		})

		t.Run("Linked", func(t *testing.T) {
			tb := NewTable()
			defer tb.Clear()

			id := uuid.Must(uuid.NewV7())

			meta1 := &spec.Meta{
				ID:        id,
				Kind:      kind,
				Namespace: resource.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}
			meta2 := &spec.Meta{
				ID:        id,
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
			meta4 := &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: resource.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}

			meta1.Ports = map[string][]spec.Port{
				node.PortOut: {
					{
						Name: meta3.GetName(),
						Port: node.PortIn,
					},
				},
			}
			meta2.Ports = map[string][]spec.Port{
				node.PortOut: {
					{
						Name: meta4.GetName(),
						Port: node.PortIn,
					},
				},
			}

			sym3 := &Symbol{Spec: meta3, Node: node.NewOneToOneNode(nil)}
			err := tb.Insert(sym3)
			assert.NoError(t, err)

			sym4 := &Symbol{Spec: meta4, Node: node.NewOneToOneNode(nil)}
			err = tb.Insert(sym4)
			assert.NoError(t, err)

			sym1 := &Symbol{Spec: meta1, Node: node.NewOneToOneNode(nil)}
			err = tb.Insert(sym1)
			assert.NoError(t, err)

			p1 := sym1.Out(node.PortOut)
			assert.Equal(t, 1, p1.Links())

			sym2 := &Symbol{Spec: meta2, Node: node.NewOneToOneNode(nil)}
			err = tb.Insert(sym2)
			assert.NoError(t, err)

			p2 := sym2.Out(node.PortOut)
			assert.Equal(t, 1, p2.Links())
		})
	})
}

func TestTable_Free(t *testing.T) {
	kind := faker.UUIDHyphenated()

	tb := NewTable()
	defer tb.Clear()

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

func TestTable_LookupByID(t *testing.T) {
	kind := faker.UUIDHyphenated()

	tb := NewTable()
	defer tb.Clear()

	meta := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: resource.DefaultNamespace,
	}

	sym := &Symbol{Spec: meta, Node: node.NewOneToOneNode(nil)}
	err := tb.Insert(sym)
	assert.NoError(t, err)

	r, ok := tb.Lookup(sym.ID())
	assert.True(t, ok)
	assert.Equal(t, sym, r)
}

func TestTable_Keys(t *testing.T) {
	kind := faker.UUIDHyphenated()

	tb := NewTable()
	defer tb.Clear()

	meta := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	sym := &Symbol{Spec: meta, Node: node.NewOneToOneNode(nil)}
	err := tb.Insert(sym)
	assert.NoError(t, err)

	ids := tb.Keys()
	assert.Contains(t, ids, sym.ID())
}

func TestTable_Hook(t *testing.T) {
	kind := faker.UUIDHyphenated()

	loaded := 0
	unloaded := 0

	tb := NewTable(TableOptions{
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
	defer tb.Clear()

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
	defer tb.Clear()

	for i := 0; i < b.N; i++ {
		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
		}

		sym := &Symbol{Spec: meta}
		_ = tb.Insert(sym)
	}
}

func BenchmarkTable_Free(b *testing.B) {
	kind := faker.UUIDHyphenated()

	tb := NewTable()
	defer tb.Clear()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
		}

		sym := &Symbol{Spec: meta}
		_ = tb.Insert(sym)

		b.StartTimer()

		_, _ = tb.Free(meta.GetID())
	}
}
