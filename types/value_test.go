package types

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestCast(t *testing.T) {
	_, err := Cast[Map](NewMap())
	require.NoError(t, err)

	_, err = Cast[Map](NewInt(0))
	require.Error(t, err)

	_, err = Cast[Map](NewMap(), errors.New(faker.Sentence()))
	require.Error(t, err)
}

func TestLookup(t *testing.T) {
	t.Run("map", func(t *testing.T) {
		m := NewMap()
		s := NewMap()

		s = s.Set(NewString("bar"), NewString("baz"))
		m = m.Set(NewString("foo"), s)

		val := Lookup(m, "foo", "bar")
		require.NotNil(t, val)
		require.Equal(t, NewString("baz"), val)
	})

	t.Run("slice", func(t *testing.T) {
		m := NewMap()
		s := NewSlice()

		s = s.Append(NewString("first"))
		s = s.Append(NewString("second"))
		m = m.Set(NewString("foo"), s)

		val := Lookup(m, "foo", 1)
		require.NotNil(t, val)
		require.Equal(t, NewString("second"), val)
	})
}
