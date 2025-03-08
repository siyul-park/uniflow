package driver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewInMemoryDriver(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	driver := NewInMemoryDriver()
	require.NotNil(t, driver)
	require.NoError(t, driver.Close(ctx))
}

func TestInMemoryDriver_SpecStore(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	driver := NewInMemoryDriver()
	defer driver.Close(ctx)

	store, err := driver.NewSpecStore(ctx, "")
	require.NoError(t, err)
	require.NotNil(t, store)
}

func TestInMemoryDriver_ValueStore(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	driver := NewInMemoryDriver()
	defer driver.Close(ctx)

	store, err := driver.NewValueStore(ctx, "")
	require.NoError(t, err)
	require.NotNil(t, store)
}
