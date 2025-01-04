package driver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMongoDriver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	driver, err := NewMongoDriver("memongodb://", "")
	assert.NoError(t, err)
	assert.NotNil(t, driver)
	assert.NoError(t, driver.Close(ctx))
}

func TestMongoDriver_SpecStore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	driver, _ := NewMongoDriver("memongodb://", "")
	defer driver.Close(ctx)

	store, err := driver.NewSpecStore(ctx, "")
	assert.NoError(t, err)
	assert.NotNil(t, store)
}

func TestMongoDriver_ValueStore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	driver, _ := NewMongoDriver("memongodb://", "")
	defer driver.Close(ctx)

	store, err := driver.NewValueStore(ctx, "")
	assert.NoError(t, err)
	assert.NotNil(t, store)
}
