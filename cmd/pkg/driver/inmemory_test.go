package driver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewInMemoryDriver(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	driver := NewInMemoryDriver()
	assert.NotNil(t, driver)
	assert.NoError(t, driver.Close(ctx))
}

func TestInMemoryDriver_SpecStore(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	driver := NewInMemoryDriver()
	defer driver.Close(ctx)

	store, err := driver.SpecStore(ctx, "")
	assert.NoError(t, err)
	assert.NotNil(t, store)
}

func TestInMemoryDriver_ValueStore(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	driver := NewInMemoryDriver()
	defer driver.Close(ctx)

	store, err := driver.ValueStore(ctx, "")
	assert.NoError(t, err)
	assert.NotNil(t, store)
}

func TestInMemoryDriver_ChartStore(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	driver := NewInMemoryDriver()
	defer driver.Close(ctx)

	store, err := driver.ChartStore(ctx, "")
	assert.NoError(t, err)
	assert.NotNil(t, store)
}
