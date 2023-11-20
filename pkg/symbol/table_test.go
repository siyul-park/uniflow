package symbol

import (
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/stretchr/testify/assert"
)

func TestTable_Insert(t *testing.T) {
	t.Run("not exists", func(t *testing.T) {
		tb := NewTable()
		defer func() { _ = tb.Close() }()

		n := node.NewOneToOneNode(node.OneToOneNodeConfig{})

		s, err := tb.Insert(n)
		assert.NoError(t, err)
		assert.NotNil(t, s)
		assert.Equal(t, n.ID(), s.ID())
	})

	t.Run("exists", func(t *testing.T) {
		tb := NewTable()
		defer func() { _ = tb.Close() }()

		id := ulid.Make()

		n1 := node.NewOneToOneNode(node.OneToOneNodeConfig{ID: id})
		n2 := node.NewOneToOneNode(node.OneToOneNodeConfig{ID: id})

		s1, err := tb.Insert(n1)
		assert.NoError(t, err)
		assert.NotNil(t, s1)
		assert.Equal(t, n1.ID(), s1.ID())

		s2, err := tb.Insert(n1)
		assert.NoError(t, err)
		assert.NotNil(t, s2)
		assert.Equal(t, n2.ID(), s2.ID())
	})
}

func TestTable_Free(t *testing.T) {
	t.Run("not exists", func(t *testing.T) {
		tb := NewTable()
		defer func() { _ = tb.Close() }()

		n := node.NewOneToOneNode(node.OneToOneNodeConfig{})

		ok, err := tb.Free(n.ID())
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("exists", func(t *testing.T) {
		tb := NewTable()
		defer func() { _ = tb.Close() }()

		n := node.NewOneToOneNode(node.OneToOneNodeConfig{})

		tb.Insert(n)

		ok, err := tb.Free(n.ID())
		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

func TestTable_Lookup(t *testing.T) {
	t.Run("not exists", func(t *testing.T) {
		tb := NewTable()
		defer func() { _ = tb.Close() }()

		n := node.NewOneToOneNode(node.OneToOneNodeConfig{})

		s, ok := tb.Lookup(n.ID())
		assert.False(t, ok)
		assert.Nil(t, s)
	})

	t.Run("exists", func(t *testing.T) {
		tb := NewTable()
		defer func() { _ = tb.Close() }()

		n := node.NewOneToOneNode(node.OneToOneNodeConfig{})

		tb.Insert(n)

		s, ok := tb.Lookup(n.ID())
		assert.True(t, ok)
		assert.NotNil(t, s)
		assert.Equal(t, n.ID(), s.ID())
	})
}
