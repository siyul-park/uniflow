package driver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewMongoDriver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	driver, err := NewMongoDriver("memongodb://", "")
	require.NoError(t, err)
	require.NotNil(t, driver)
	require.NoError(t, driver.Close(ctx))
}

func TestMongoDriver_SpecStore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	driver, _ := NewMongoDriver("memongodb://", "")
	defer driver.Close(ctx)

	store, err := driver.NewSpecStore(ctx, "")
	require.NoError(t, err)
	require.NotNil(t, store)
}

func TestMongoDriver_ValueStore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	driver, _ := NewMongoDriver("memongodb://", "")
	defer driver.Close(ctx)

	store, err := driver.NewValueStore(ctx, "")
	require.NoError(t, err)
	require.NotNil(t, store)
}
