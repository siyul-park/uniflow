package symbol

import (
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/assert"
)

func TestTable_Insert(t *testing.T) {
	t.Run("not exists", func(t *testing.T) {
		tb := NewTable()
		defer tb.Close()

		n1 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
		defer n1.Close()
		n2 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
		defer n2.Close()
		n3 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
		defer n3.Close()

		spec1 := &scheme.SpecMeta{
			ID:        n1.ID(),
			Namespace: scheme.NamespaceDefault,
			Links: map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						ID:   n2.ID(),
						Port: node.PortIn,
					},
				},
			},
		}
		spec2 := &scheme.SpecMeta{
			ID:        n2.ID(),
			Namespace: scheme.NamespaceDefault,
			Links: map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						ID:   n3.ID(),
						Port: node.PortIn,
					},
				},
			},
		}
		spec3 := &scheme.SpecMeta{
			ID:        n3.ID(),
			Namespace: scheme.NamespaceDefault,
			Links: map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						ID:   n1.ID(),
						Port: node.PortIn,
					},
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

	t.Run("exists", func(t *testing.T) {
		tb := NewTable()
		defer tb.Close()

		id := ulid.Make()

		n1 := node.NewOneToOneNode(node.OneToOneNodeConfig{ID: id})
		defer n1.Close()
		n2 := node.NewOneToOneNode(node.OneToOneNodeConfig{ID: id})
		defer n2.Close()
		n3 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
		defer n3.Close()
		n4 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
		defer n3.Close()

		spec1 := &scheme.SpecMeta{
			ID:        id,
			Namespace: scheme.NamespaceDefault,
			Links: map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						ID:   n3.ID(),
						Port: node.PortIn,
					},
				},
			},
		}
		spec2 := &scheme.SpecMeta{
			ID:        id,
			Namespace: scheme.NamespaceDefault,
			Links: map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						ID:   n4.ID(),
						Port: node.PortIn,
					},
				},
			},
		}
		spec3 := &scheme.SpecMeta{
			ID:        n3.ID(),
			Namespace: scheme.NamespaceDefault,
			Links: map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						ID:   id,
						Port: node.PortIn,
					},
				},
			},
		}
		spec4 := &scheme.SpecMeta{
			ID:        n4.ID(),
			Namespace: scheme.NamespaceDefault,
			Links: map[string][]scheme.PortLocation{
				node.PortOut: {
					{
						ID:   id,
						Port: node.PortIn,
					},
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
		ID:        n1.ID(),
		Namespace: scheme.NamespaceDefault,
		Links: map[string][]scheme.PortLocation{
			node.PortOut: {
				{
					ID:   n2.ID(),
					Port: node.PortIn,
				},
			},
		},
	}
	spec2 := &scheme.SpecMeta{
		ID:        n2.ID(),
		Namespace: scheme.NamespaceDefault,
		Links: map[string][]scheme.PortLocation{
			node.PortOut: {
				{
					ID:   n3.ID(),
					Port: node.PortIn,
				},
			},
		},
	}
	spec3 := &scheme.SpecMeta{
		ID:        n3.ID(),
		Namespace: scheme.NamespaceDefault,
		Links: map[string][]scheme.PortLocation{
			node.PortOut: {
				{
					ID:   n1.ID(),
					Port: node.PortIn,
				},
			},
		},
	}

	p1, _ := n1.Port(node.PortIn)
	p2, _ := n2.Port(node.PortIn)
	p3, _ := n3.Port(node.PortIn)

	_ = tb.Insert(&Symbol{Node: n1, Spec: spec1})
	_ = tb.Insert(&Symbol{Node: n2, Spec: spec2})
	_ = tb.Insert(&Symbol{Node: n3, Spec: spec3})

	ok, err := tb.Free(n1.ID())
	assert.NoError(t, err)
	assert.True(t, ok)

	assert.Equal(t, 0, p1.Links())

	ok, err = tb.Free(n2.ID())
	assert.NoError(t, err)
	assert.True(t, ok)

	assert.Equal(t, 0, p1.Links())
	assert.Equal(t, 0, p2.Links())

	ok, err = tb.Free(n3.ID())
	assert.NoError(t, err)
	assert.True(t, ok)

	assert.Equal(t, 0, p1.Links())
	assert.Equal(t, 0, p2.Links())
	assert.Equal(t, 0, p3.Links())
}

func TestTable_Lookup(t *testing.T) {
	tb := NewTable()
	defer tb.Close()

	n := node.NewOneToOneNode(node.OneToOneNodeConfig{})
	defer n.Close()
	spec := &scheme.SpecMeta{
		ID: n.ID(),
	}
	sym := &Symbol{Node: n, Spec: spec}

	_ = tb.Insert(sym)

	r, ok := tb.Lookup(n.ID())
	assert.True(t, ok)
	assert.Equal(t, sym, r)
}
