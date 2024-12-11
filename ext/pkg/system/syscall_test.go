package system

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/stretchr/testify/assert"
)

func TestWatchResource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	st := resource.NewStore[resource.Resource]()
	fn := WatchResource(st)

	_, err := fn(ctx)
	assert.NoError(t, err)
}

func TestCreateResource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	st := resource.NewStore[resource.Resource]()
	fn := CreateResource(st)

	meta := &resource.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Name: faker.Word(),
	}

	res, err := fn(ctx, []resource.Resource{meta})
	assert.NoError(t, err)
	assert.Len(t, res, 1)
}

func TestReadResource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	st := resource.NewStore[resource.Resource]()
	fn := ReadResource(st)

	meta := &resource.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Name: faker.Word(),
	}

	_, err := st.Store(ctx, meta)
	assert.NoError(t, err)

	res, err := fn(ctx, []resource.Resource{meta})
	assert.NoError(t, err)
	assert.Len(t, res, 1)
}

func TestUpdateResource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	st := resource.NewStore[resource.Resource]()
	fn := UpdateResource(st)

	meta := &resource.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Name: faker.Word(),
	}

	_, err := st.Store(ctx, meta)
	assert.NoError(t, err)

	res, err := fn(ctx, []resource.Resource{meta})
	assert.NoError(t, err)
	assert.Len(t, res, 1)
}

func TestDeleteResource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	st := resource.NewStore[resource.Resource]()
	fn := DeleteResource(st)

	meta := &resource.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Name: faker.Word(),
	}

	_, err := st.Store(ctx, meta)
	assert.NoError(t, err)

	res, err := fn(ctx, []resource.Resource{meta})
	assert.NoError(t, err)
	assert.Len(t, res, 1)
}
