package driver

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMongoDriver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	driver, err := NewMongoDriver(ctx, "memongodb://", "")
	assert.NoError(t, err)
	assert.NotNil(t, driver)
	assert.NoError(t, driver.Close(ctx))
}

func TestMongoDriver_SpecStore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	driver, _ := NewMongoDriver(ctx, "memongodb://", "")
	defer driver.Close(ctx)

	store, err := driver.SpecStore(ctx, "")
	assert.NoError(t, err)
	assert.NotNil(t, store)
}

func TestMongoDriver_SecretStore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	driver, _ := NewMongoDriver(ctx, "memongodb://", "")
	defer driver.Close(ctx)

	store, err := driver.SecretStore(ctx, "")
	assert.NoError(t, err)
	assert.NotNil(t, store)
}

func TestMongoDriver_ChartStore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	driver, _ := NewMongoDriver(ctx, "memongodb://", "")
	defer driver.Close(ctx)

	store, err := driver.ChartStore(ctx, "")
	assert.NoError(t, err)
	assert.NotNil(t, store)
}
