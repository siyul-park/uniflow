package symbol

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/assert"
)

func TestTable_Insert(t *testing.T) {
	tb := NewTable()
	defer tb.Close()

	n1 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
	defer n1.Close()
	n2 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
	defer n2.Close()
	n3 := node.NewOneToOneNode(node.OneToOneNodeConfig{})
	defer n3.Close()

	spec1 := &scheme.SpecMeta{
		ID: n1.ID(),
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
		ID: n2.ID(),
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
		ID: n3.ID(),
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

	err := tb.Insert(n1, spec1)
	assert.NoError(t, err)

	assert.Equal(t, 0, p1.Links())
	assert.Equal(t, 0, p2.Links())
	assert.Equal(t, 0, p3.Links())

	err = tb.Insert(n2, spec2)
	assert.NoError(t, err)

	assert.Equal(t, 0, p1.Links())
	assert.Equal(t, 1, p2.Links())
	assert.Equal(t, 0, p3.Links())

	err = tb.Insert(n3, spec3)
	assert.NoError(t, err)

	assert.Equal(t, 1, p1.Links())
	assert.Equal(t, 1, p2.Links())
	assert.Equal(t, 1, p3.Links())
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
		ID: n1.ID(),
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
		ID: n2.ID(),
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
		ID: n3.ID(),
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

	_ = tb.Insert(n1, spec1)
	_ = tb.Insert(n2, spec2)
	_ = tb.Insert(n3, spec3)

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

	_ = tb.Insert(n, spec)

	r, ok := tb.Lookup(n.ID())
	assert.True(t, ok)
	assert.Equal(t, r, n)
}
