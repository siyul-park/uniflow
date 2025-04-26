package driver

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/pkg/types"
)

func TestCursor_All(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	doc := types.NewMap(
		types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	c := newCursor([]types.Map{doc})
	defer c.Close(ctx)

	var docs any
	err := c.All(ctx, &docs)
	require.NoError(t, err)
	require.Len(t, docs, 1)
}

func TestCursor_Next(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	doc := types.NewMap(
		types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	c := newCursor([]types.Map{doc})
	defer c.Close(ctx)

	ok := c.Next(ctx)
	require.True(t, ok)

	ok = c.Next(ctx)
	require.False(t, ok)
}

func TestCursor_Decode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	doc := types.NewMap(
		types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	c := newCursor([]types.Map{doc})
	defer c.Close(ctx)

	ok := c.Next(ctx)
	require.True(t, ok)

	var val types.Value
	err := c.Decode(&val)
	require.NoError(t, err)
	require.Equal(t, doc, val)
}
