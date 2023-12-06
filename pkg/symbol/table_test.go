package symbol

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/assert"
)

func TestTable_Insert(t *testing.T) {
	t.Run("by ID", func(t *testing.T) {
		t.Run("when not exists", func(t *testing.T) {
			tb := NewTable()
			defer tb.Close()

			n1 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
			defer n1.Close()
			n2 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
			defer n2.Close()
			n3 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
			defer n3.Close()

			spec1 := &scheme.SpecMeta{
				ID:        ulid.Make(),
				Namespace: scheme.DefaultNamespace,
			}
			spec2 := &scheme.SpecMeta{
				ID:        ulid.Make(),
				Namespace: scheme.DefaultNamespace,
			}
			spec3 := &scheme.SpecMeta{
				ID:        ulid.Make(),
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

			p1, _ := n1.Port(node.PortIn)
			p2, _ := n2.Port(node.PortIn)
			p3, _ := n3.Port(node.PortIn)

			err := tb.Insert(&Symbol{Node: n1, Spec: spec1})
			assert.NoError(t, err)

			assert.Equal(t, 0, p1.Links())
			assert.Equal(t, 0, p2.Links())
			assert.Equal(t, 0, p3.Links())

			err = tb.Insert(&Symbol{Node: n2, Spec: spec2})
			assert.NoError(t, err)

			assert.Equal(t, 0, p1.Links())
			assert.Equal(t, 1, p2.Links())
			assert.Equal(t, 0, p3.Links())

			err = tb.Insert(&Symbol{Node: n3, Spec: spec3})
			assert.NoError(t, err)

			assert.Equal(t, 1, p1.Links())
			assert.Equal(t, 1, p2.Links())
			assert.Equal(t, 1, p3.Links())
		})

		t.Run("when exists", func(t *testing.T) {
			tb := NewTable()
			defer tb.Close()

			id := ulid.Make()

			n1 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
			defer n1.Close()
			n2 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
			defer n2.Close()
			n3 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
			defer n3.Close()
			n4 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
			defer n3.Close()

			spec1 := &scheme.SpecMeta{
				ID:        id,
				Namespace: scheme.DefaultNamespace,
			}
			spec2 := &scheme.SpecMeta{
				ID:        id,
				Namespace: scheme.DefaultNamespace,
			}
			spec3 := &scheme.SpecMeta{
				ID:        ulid.Make(),
				Namespace: scheme.DefaultNamespace,
			}
			spec4 := &scheme.SpecMeta{
				ID:        ulid.Make(),
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
			spec3.Links = map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						ID:   id,
						Port: node.PortIn,
					},
				},
			}
			spec4.Links = map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						ID:   id,
						Port: node.PortIn,
					},
				},
			}

			p1, _ := n1.Port(node.PortIn)
			p2, _ := n2.Port(node.PortIn)
			p3, _ := n3.Port(node.PortIn)
			p4, _ := n4.Port(node.PortIn)

			_ = tb.Insert(&Symbol{Node: n3, Spec: spec3})
			_ = tb.Insert(&Symbol{Node: n4, Spec: spec4})

			err := tb.Insert(&Symbol{Node: n1, Spec: spec1})
			assert.NoError(t, err)

			assert.Equal(t, 2, p1.Links())
			assert.Equal(t, 0, p2.Links())
			assert.Equal(t, 1, p3.Links())

			err = tb.Insert(&Symbol{Node: n2, Spec: spec2})
			assert.NoError(t, err)

			assert.Equal(t, 0, p1.Links())
			assert.Equal(t, 2, p2.Links())
			assert.Equal(t, 1, p4.Links())
		})
	})

	t.Run("by name", func(t *testing.T) {
		t.Run("when not exists", func(t *testing.T) {
			tb := NewTable()
			defer tb.Close()

			n1 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
			defer n1.Close()
			n2 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
			defer n2.Close()
			n3 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
			defer n3.Close()

			spec1 := &scheme.SpecMeta{
				ID:        ulid.Make(),
				Namespace: scheme.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}
			spec2 := &scheme.SpecMeta{
				ID:        ulid.Make(),
				Namespace: scheme.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}
			spec3 := &scheme.SpecMeta{
				ID:        ulid.Make(),
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

			p1, _ := n1.Port(node.PortIn)
			p2, _ := n2.Port(node.PortIn)
			p3, _ := n3.Port(node.PortIn)

			err := tb.Insert(&Symbol{Node: n1, Spec: spec1})
			assert.NoError(t, err)

			assert.Equal(t, 0, p1.Links())
			assert.Equal(t, 0, p2.Links())
			assert.Equal(t, 0, p3.Links())

			err = tb.Insert(&Symbol{Node: n2, Spec: spec2})
			assert.NoError(t, err)

			assert.Equal(t, 0, p1.Links())
			assert.Equal(t, 1, p2.Links())
			assert.Equal(t, 0, p3.Links())

			err = tb.Insert(&Symbol{Node: n3, Spec: spec3})
			assert.NoError(t, err)

			assert.Equal(t, 1, p1.Links())
			assert.Equal(t, 1, p2.Links())
			assert.Equal(t, 1, p3.Links())
		})

		t.Run("when exists", func(t *testing.T) {
			tb := NewTable()
			defer tb.Close()

			id := ulid.Make()
			name := faker.UUIDHyphenated()

			n1 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
			defer n1.Close()
			n2 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
			defer n2.Close()
			n3 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
			defer n3.Close()
			n4 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
			defer n3.Close()

			spec1 := &scheme.SpecMeta{
				ID:        id,
				Namespace: scheme.DefaultNamespace,
				Name:      name,
			}
			spec2 := &scheme.SpecMeta{
				ID:        id,
				Namespace: scheme.DefaultNamespace,
				Name:      name,
			}
			spec3 := &scheme.SpecMeta{
				ID:        ulid.Make(),
				Namespace: scheme.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			}
			spec4 := &scheme.SpecMeta{
				ID:        ulid.Make(),
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
			spec3.Links = map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						Name: name,
						Port: node.PortIn,
					},
				},
			}
			spec4.Links = map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						Name: name,
						Port: node.PortIn,
					},
				},
			}

			p1, _ := n1.Port(node.PortIn)
			p2, _ := n2.Port(node.PortIn)
			p3, _ := n3.Port(node.PortIn)
			p4, _ := n4.Port(node.PortIn)

			_ = tb.Insert(&Symbol{Node: n3, Spec: spec3})
			_ = tb.Insert(&Symbol{Node: n4, Spec: spec4})

			err := tb.Insert(&Symbol{Node: n1, Spec: spec1})
			assert.NoError(t, err)

			assert.Equal(t, 2, p1.Links())
			assert.Equal(t, 0, p2.Links())
			assert.Equal(t, 1, p3.Links())

			err = tb.Insert(&Symbol{Node: n2, Spec: spec2})
			assert.NoError(t, err)

			assert.Equal(t, 0, p1.Links())
			assert.Equal(t, 2, p2.Links())
			assert.Equal(t, 1, p4.Links())
		})
	})
}

func TestTable_Free(t *testing.T) {
	tb := NewTable()
	defer tb.Close()

	n1 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
	defer n1.Close()
	n2 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
	defer n2.Close()
	n3 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
	defer n3.Close()

	spec1 := &scheme.SpecMeta{
		ID:        ulid.Make(),
		Namespace: scheme.DefaultNamespace,
	}
	spec2 := &scheme.SpecMeta{
		ID:        ulid.Make(),
		Namespace: scheme.DefaultNamespace,
	}
	spec3 := &scheme.SpecMeta{
		ID:        ulid.Make(),
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

	p1, _ := n1.Port(node.PortIn)
	p2, _ := n2.Port(node.PortIn)
	p3, _ := n3.Port(node.PortIn)

	_ = tb.Insert(&Symbol{Node: n1, Spec: spec1})
	_ = tb.Insert(&Symbol{Node: n2, Spec: spec2})
	_ = tb.Insert(&Symbol{Node: n3, Spec: spec3})

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
	tb := NewTable()
	defer tb.Close()

	n := node.NewOneToOneNode(node.OneToOneNodeConfig{})
	defer n.Close()
	spec := &scheme.SpecMeta{
		ID: ulid.Make(),
	}
	sym := &Symbol{Node: n, Spec: spec}

	_ = tb.Insert(sym)

	r, ok := tb.LookupByID(spec.GetID())
	assert.True(t, ok)
	assert.Equal(t, sym, r)
}

func TestTable_LookupByName(t *testing.T) {
	tb := NewTable()
	defer tb.Close()

	n := node.NewOneToOneNode(node.OneToOneNodeConfig{})
	defer n.Close()
	spec := &scheme.SpecMeta{
		ID:        ulid.Make(),
		Namespace: scheme.DefaultNamespace,
		Name:      faker.Word(),
	}
	sym := &Symbol{Node: n, Spec: spec}

	_ = tb.Insert(sym)

	r, ok := tb.LookupByName(spec.GetNamespace(), spec.GetName())
	assert.True(t, ok)
	assert.Equal(t, sym, r)
}
