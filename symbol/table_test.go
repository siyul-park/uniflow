package symbol

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/node"
	"github.com/siyul-park/uniflow/scheme"
	"github.com/siyul-park/uniflow/spec"
	"github.com/stretchr/testify/assert"
)

func TestTable_Insert(t *testing.T) {
	s := scheme.New()

	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	t.Run("Link By ID", func(t *testing.T) {
		t.Run("Unlinked", func(t *testing.T) {
			tb := NewTable(s)
			defer tb.Clear()

			meta1 := &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: spec.DefaultNamespace,
			}
			meta2 := &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: spec.DefaultNamespace,
			}
			meta3 := &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: spec.DefaultNamespace,
			}

			meta1.Links = map[string][]spec.PortLocation{
				node.PortOut: {
					{
						ID:   meta2.GetID(),
						Port: node.PortIn,
					},
				},
			}
			meta2.Links = map[string][]spec.PortLocation{
				node.PortOut: {
					{
						ID:   meta3.GetID(),
						Port: node.PortIn,
					},
				},
			}
			meta3.Links = map[string][]spec.PortLocation{
				node.PortOut: {
					{
						ID:   meta1.GetID(),
						Port: node.PortIn,
					},
				},
			}

			sym1, err := tb.Insert(meta1)
			assert.NoError(t, err)

			sym2, err := tb.Insert(meta2)
			assert.NoError(t, err)

			sym3, err := tb.Insert(meta3)
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

			meta1 := &spec.Meta{
				ID:        id,
				Kind:      kind,
				Namespace: spec.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}
			meta2 := &spec.Meta{
				ID:        id,
				Kind:      kind,
				Namespace: spec.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}
			meta3 := &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: spec.DefaultNamespace,
			}
			meta4 := &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: spec.DefaultNamespace,
			}

			meta1.Links = map[string][]spec.PortLocation{
				node.PortOut: {
					{
						ID:   meta3.GetID(),
						Port: node.PortIn,
					},
				},
			}
			meta2.Links = map[string][]spec.PortLocation{
				node.PortOut: {
					{
						ID:   meta4.GetID(),
						Port: node.PortIn,
					},
				},
			}

			_, err := tb.Insert(meta3)
			assert.NoError(t, err)

			_, err = tb.Insert(meta4)
			assert.NoError(t, err)

			sym1, err := tb.Insert(meta1)
			assert.NoError(t, err)

			p1 := sym1.Out(node.PortOut)
			assert.Equal(t, 1, p1.Links())

			sym2, err := tb.Insert(meta2)
			assert.NoError(t, err)

			p2 := sym2.Out(node.PortOut)
			assert.Equal(t, 1, p2.Links())
		})
	})

	t.Run("Link By Name", func(t *testing.T) {
		t.Run("Unlinked", func(t *testing.T) {
			tb := NewTable(s)
			defer tb.Clear()

			meta1 := &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: spec.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}
			meta2 := &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: spec.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}
			meta3 := &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: spec.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}

			meta1.Links = map[string][]spec.PortLocation{
				node.PortOut: {
					{
						Name: meta2.GetName(),
						Port: node.PortIn,
					},
				},
			}
			meta2.Links = map[string][]spec.PortLocation{
				node.PortOut: {
					{
						Name: meta3.GetName(),
						Port: node.PortIn,
					},
				},
			}
			meta3.Links = map[string][]spec.PortLocation{
				node.PortOut: {
					{
						Name: meta1.GetName(),
						Port: node.PortIn,
					},
				},
			}

			sym1, err := tb.Insert(meta1)
			assert.NoError(t, err)

			sym2, err := tb.Insert(meta2)
			assert.NoError(t, err)

			sym3, err := tb.Insert(meta3)
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

			meta1 := &spec.Meta{
				ID:        id,
				Kind:      kind,
				Namespace: spec.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}
			meta2 := &spec.Meta{
				ID:        id,
				Kind:      kind,
				Namespace: spec.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}
			meta3 := &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: spec.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}
			meta4 := &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: spec.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}

			meta1.Links = map[string][]spec.PortLocation{
				node.PortOut: {
					{
						Name: meta3.GetName(),
						Port: node.PortIn,
					},
				},
			}
			meta2.Links = map[string][]spec.PortLocation{
				node.PortOut: {
					{
						Name: meta4.GetName(),
						Port: node.PortIn,
					},
				},
			}

			_, err := tb.Insert(meta3)
			assert.NoError(t, err)

			_, err = tb.Insert(meta4)
			assert.NoError(t, err)

			sym1, err := tb.Insert(meta1)
			assert.NoError(t, err)

			p1 := sym1.Out(node.PortOut)
			assert.Equal(t, 1, p1.Links())

			sym2, err := tb.Insert(meta2)
			assert.NoError(t, err)

			p2 := sym2.Out(node.PortOut)
			assert.Equal(t, 1, p2.Links())
		})
	})
}

func TestTable_Free(t *testing.T) {
	s := scheme.New()

	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	tb := NewTable(s)
	defer tb.Clear()

	meta1 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: spec.DefaultNamespace,
	}
	meta2 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: spec.DefaultNamespace,
	}
	meta3 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: spec.DefaultNamespace,
	}

	meta1.Links = map[string][]spec.PortLocation{
		node.PortOut: {
			{
				ID:   meta2.GetID(),
				Port: node.PortIn,
			},
		},
	}
	meta2.Links = map[string][]spec.PortLocation{
		node.PortOut: {
			{
				ID:   meta3.GetID(),
				Port: node.PortIn,
			},
		},
	}
	meta3.Links = map[string][]spec.PortLocation{
		node.PortOut: {
			{
				ID:   meta1.GetID(),
				Port: node.PortIn,
			},
		},
	}

	sym1, err := tb.Insert(meta1)
	assert.NoError(t, err)

	sym2, err := tb.Insert(meta2)
	assert.NoError(t, err)

	sym3, err := tb.Insert(meta3)
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
	s := scheme.New()

	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	tb := NewTable(s)
	defer tb.Clear()

	spec := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: spec.DefaultNamespace,
	}

	sym, _ := tb.Insert(spec)

	r, ok := tb.LookupByID(sym.ID())
	assert.True(t, ok)
	assert.Equal(t, sym, r)
}

func TestTable_LookupByName(t *testing.T) {
	s := scheme.New()

	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	tb := NewTable(s)
	defer tb.Clear()

	meta := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: spec.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	sym, _ := tb.Insert(meta)

	r, ok := tb.LookupByName(sym.Namespace(), sym.Name())
	assert.True(t, ok)
	assert.Equal(t, sym, r)
}

func TestTable_Keys(t *testing.T) {
	s := scheme.New()

	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	tb := NewTable(s)
	defer tb.Clear()

	meta := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: spec.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	sym, _ := tb.Insert(meta)

	ids := tb.Keys()
	assert.Contains(t, ids, sym.ID())
}

func TestTable_Hook(t *testing.T) {
	s := scheme.New()

	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
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

	meta1 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: spec.DefaultNamespace,
	}
	meta2 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: spec.DefaultNamespace,
	}
	meta3 := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: spec.DefaultNamespace,
	}

	meta1.Links = map[string][]spec.PortLocation{
		node.PortOut: {
			{
				ID:   meta2.GetID(),
				Port: node.PortIn,
			},
		},
	}
	meta2.Links = map[string][]spec.PortLocation{
		node.PortOut: {
			{
				ID:   meta3.GetID(),
				Port: node.PortIn,
			},
		},
	}
	meta3.Links = map[string][]spec.PortLocation{
		node.PortOut: {
			{
				ID:   meta1.GetID(),
				Port: node.PortIn,
			},
		},
	}

	sym1, err := tb.Insert(meta1)
	assert.NoError(t, err)
	assert.Equal(t, 0, loaded)
	assert.Equal(t, 0, unloaded)

	sym2, err := tb.Insert(meta2)
	assert.NoError(t, err)
	assert.Equal(t, 0, loaded)
	assert.Equal(t, 0, unloaded)

	sym3, err := tb.Insert(meta3)
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

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	tb := NewTable(s)
	defer tb.Clear()

	for i := 0; i < b.N; i++ {
		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
		}

		_, _ = tb.Insert(meta)
	}
}

func BenchmarkTable_Free(b *testing.B) {
	s := scheme.New()

	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	tb := NewTable(s)
	defer tb.Clear()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: spec.DefaultNamespace,
		}

		_, _ = tb.Insert(meta)

		b.StartTimer()

		_, _ = tb.Free(meta.GetID())
	}
}
