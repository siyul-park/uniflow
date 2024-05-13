package symbol

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/assert"
)

func TestTable_Insert(t *testing.T) {
	s := scheme.New()

	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	t.Run("Link By ID", func(t *testing.T) {
		t.Run("Unlinked", func(t *testing.T) {
			tb := NewTable(s)
			defer tb.Clear()

			spec1 := &scheme.SpecMeta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: scheme.DefaultNamespace,
			}
			spec2 := &scheme.SpecMeta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: scheme.DefaultNamespace,
			}
			spec3 := &scheme.SpecMeta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: scheme.DefaultNamespace,
			}

			spec1.Links = map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						ID:   spec2.GetID(),
						Port: node.PortIn,
					},
				},
			}
			spec2.Links = map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						ID:   spec3.GetID(),
						Port: node.PortIn,
					},
				},
			}
			spec3.Links = map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						ID:   spec1.GetID(),
						Port: node.PortIn,
					},
				},
			}

			sym1, err := tb.Insert(spec1)
			assert.NoError(t, err)

			sym2, err := tb.Insert(spec2)
			assert.NoError(t, err)

			sym3, err := tb.Insert(spec3)
			assert.NoError(t, err)

			p1 := sym1.Out(node.PortOut)
			p2 := sym2.Out(node.PortOut)
			p3 := sym3.Out(node.PortOut)

			assert.Equal(t, 1, p1.Links())
			assert.Equal(t, 1, p2.Links())
			assert.Equal(t, 1, p3.Links())
		})

		t.Run("Linked", func(t *testing.T) {
			tb := NewTable(s)
			defer tb.Clear()

			id := uuid.Must(uuid.NewV7())

			spec1 := &scheme.SpecMeta{
				ID:        id,
				Kind:      kind,
				Namespace: scheme.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}
			spec2 := &scheme.SpecMeta{
				ID:        id,
				Kind:      kind,
				Namespace: scheme.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}
			spec3 := &scheme.SpecMeta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: scheme.DefaultNamespace,
			}
			spec4 := &scheme.SpecMeta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: scheme.DefaultNamespace,
			}

			spec1.Links = map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						ID:   spec3.GetID(),
						Port: node.PortIn,
					},
				},
			}
			spec2.Links = map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						ID:   spec4.GetID(),
						Port: node.PortIn,
					},
				},
			}

			_, err := tb.Insert(spec3)
			assert.NoError(t, err)

			_, err = tb.Insert(spec4)
			assert.NoError(t, err)

			sym1, err := tb.Insert(spec1)
			assert.NoError(t, err)

			p1 := sym1.Out(node.PortOut)
			assert.Equal(t, 1, p1.Links())

			sym2, err := tb.Insert(spec2)
			assert.NoError(t, err)

			p2 := sym2.Out(node.PortOut)
			assert.Equal(t, 1, p2.Links())
		})
	})

	t.Run("Link By Name", func(t *testing.T) {
		t.Run("Unlinked", func(t *testing.T) {
			tb := NewTable(s)
			defer tb.Clear()

			spec1 := &scheme.SpecMeta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: scheme.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}
			spec2 := &scheme.SpecMeta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: scheme.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}
			spec3 := &scheme.SpecMeta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: scheme.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}

			spec1.Links = map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						Name: spec2.GetName(),
						Port: node.PortIn,
					},
				},
			}
			spec2.Links = map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						Name: spec3.GetName(),
						Port: node.PortIn,
					},
				},
			}
			spec3.Links = map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						Name: spec1.GetName(),
						Port: node.PortIn,
					},
				},
			}

			sym1, err := tb.Insert(spec1)
			assert.NoError(t, err)

			sym2, err := tb.Insert(spec2)
			assert.NoError(t, err)

			sym3, err := tb.Insert(spec3)
			assert.NoError(t, err)

			p1 := sym1.Out(node.PortOut)
			p2 := sym2.Out(node.PortOut)
			p3 := sym3.Out(node.PortOut)

			assert.Equal(t, 1, p1.Links())
			assert.Equal(t, 1, p2.Links())
			assert.Equal(t, 1, p3.Links())
		})

		t.Run("Linked", func(t *testing.T) {
			tb := NewTable(s)
			defer tb.Clear()

			id := uuid.Must(uuid.NewV7())

			spec1 := &scheme.SpecMeta{
				ID:        id,
				Kind:      kind,
				Namespace: scheme.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}
			spec2 := &scheme.SpecMeta{
				ID:        id,
				Kind:      kind,
				Namespace: scheme.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}
			spec3 := &scheme.SpecMeta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: scheme.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}
			spec4 := &scheme.SpecMeta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: scheme.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}

			spec1.Links = map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						Name: spec3.GetName(),
						Port: node.PortIn,
					},
				},
			}
			spec2.Links = map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						Name: spec4.GetName(),
						Port: node.PortIn,
					},
				},
			}

			_, err := tb.Insert(spec3)
			assert.NoError(t, err)

			_, err = tb.Insert(spec4)
			assert.NoError(t, err)

			sym1, err := tb.Insert(spec1)
			assert.NoError(t, err)

			p1 := sym1.Out(node.PortOut)
			assert.Equal(t, 1, p1.Links())

			sym2, err := tb.Insert(spec2)
			assert.NoError(t, err)

			p2 := sym2.Out(node.PortOut)
			assert.Equal(t, 1, p2.Links())
		})
	})
}

func TestTable_Free(t *testing.T) {
	s := scheme.New()

	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	tb := NewTable(s)
	defer tb.Clear()

	spec1 := &scheme.SpecMeta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
	}
	spec2 := &scheme.SpecMeta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
	}
	spec3 := &scheme.SpecMeta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
	}

	spec1.Links = map[string][]scheme.PortLocation{
		node.PortOut: {
			{
				ID:   spec2.GetID(),
				Port: node.PortIn,
			},
		},
	}
	spec2.Links = map[string][]scheme.PortLocation{
		node.PortOut: {
			{
				ID:   spec3.GetID(),
				Port: node.PortIn,
			},
		},
	}
	spec3.Links = map[string][]scheme.PortLocation{
		node.PortOut: {
			{
				ID:   spec1.GetID(),
				Port: node.PortIn,
			},
		},
	}

	sym1, err := tb.Insert(spec1)
	assert.NoError(t, err)

	sym2, err := tb.Insert(spec2)
	assert.NoError(t, err)

	sym3, err := tb.Insert(spec3)
	assert.NoError(t, err)

	p1 := sym1.Out(node.PortOut)
	p2 := sym2.Out(node.PortOut)
	p3 := sym3.Out(node.PortOut)

	assert.Equal(t, 1, p1.Links())
	assert.Equal(t, 1, p2.Links())
	assert.Equal(t, 1, p3.Links())

	ok, err := tb.Free(spec1.GetID())
	assert.NoError(t, err)
	assert.True(t, ok)

	assert.Equal(t, 0, p1.Links())

	ok, err = tb.Free(spec2.GetID())
	assert.NoError(t, err)
	assert.True(t, ok)

	assert.Equal(t, 0, p1.Links())
	assert.Equal(t, 0, p2.Links())

	ok, err = tb.Free(spec3.GetID())
	assert.NoError(t, err)
	assert.True(t, ok)

	assert.Equal(t, 0, p1.Links())
	assert.Equal(t, 0, p2.Links())
	assert.Equal(t, 0, p3.Links())
}

func TestTable_LookupByID(t *testing.T) {
	s := scheme.New()

	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	tb := NewTable(s)
	defer tb.Clear()

	spec := &scheme.SpecMeta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
	}

	sym, _ := tb.Insert(spec)

	r, ok := tb.LookupByID(sym.ID())
	assert.True(t, ok)
	assert.Equal(t, sym, r)
}

func TestTable_LookupByName(t *testing.T) {
	s := scheme.New()

	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	tb := NewTable(s)
	defer tb.Clear()

	spec := &scheme.SpecMeta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	sym, _ := tb.Insert(spec)

	r, ok := tb.LookupByName(sym.Namespace(), sym.Name())
	assert.True(t, ok)
	assert.Equal(t, sym, r)
}

func TestTable_Keys(t *testing.T) {
	s := scheme.New()

	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	tb := NewTable(s)
	defer tb.Clear()

	spec := &scheme.SpecMeta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	sym, _ := tb.Insert(spec)

	ids := tb.Keys()
	assert.Contains(t, ids, sym.ID())
}

func TestTable_Hook(t *testing.T) {
	s := scheme.New()

	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	loaded := 0
	unloaded := 0

	tb := NewTable(s, TableOptions{
		LoadHooks: []LoadHook{
			LoadHookFunc(func(_ *Symbol) error {
				loaded += 1
				return nil
			}),
		},
		UnloadHooks: []UnloadHook{
			UnloadHookFunc(func(_ *Symbol) error {
				unloaded += 1
				return nil
			}),
		},
	})
	defer tb.Clear()

	spec1 := &scheme.SpecMeta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
	}
	spec2 := &scheme.SpecMeta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
	}
	spec3 := &scheme.SpecMeta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
	}

	spec1.Links = map[string][]scheme.PortLocation{
		node.PortOut: {
			{
				ID:   spec2.GetID(),
				Port: node.PortIn,
			},
		},
	}
	spec2.Links = map[string][]scheme.PortLocation{
		node.PortOut: {
			{
				ID:   spec3.GetID(),
				Port: node.PortIn,
			},
		},
	}
	spec3.Links = map[string][]scheme.PortLocation{
		node.PortOut: {
			{
				ID:   spec1.GetID(),
				Port: node.PortIn,
			},
		},
	}

	sym1, err := tb.Insert(spec1)
	assert.NoError(t, err)
	assert.Equal(t, 0, loaded)
	assert.Equal(t, 0, unloaded)

	sym2, err := tb.Insert(spec2)
	assert.NoError(t, err)
	assert.Equal(t, 0, loaded)
	assert.Equal(t, 0, unloaded)

	sym3, err := tb.Insert(spec3)
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
	s := scheme.New()

	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	tb := NewTable(s)
	defer tb.Clear()

	for i := 0; i < b.N; i++ {
		spec := &scheme.SpecMeta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
		}

		_, _ = tb.Insert(spec)
	}
}

func BenchmarkTable_Free(b *testing.B) {
	s := scheme.New()

	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	tb := NewTable(s)
	defer tb.Clear()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		spec := &scheme.SpecMeta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: scheme.DefaultNamespace,
		}

		_, _ = tb.Insert(spec)

		b.StartTimer()

		_, _ = tb.Free(spec.GetID())
	}
}
