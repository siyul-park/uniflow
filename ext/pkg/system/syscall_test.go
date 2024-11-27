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

func TestCreateResource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	st := resource.NewStore[*resource.Meta]()
	fn := CreateResource(st)

	meta := &resource.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Name: faker.Word(),
	}

	res, err := fn(ctx, []*resource.Meta{meta})
	assert.NoError(t, err)
	assert.Len(t, res, 1)
}

func TestReadResource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	st := resource.NewStore[*resource.Meta]()
	fn := ReadResource(st)

	meta := &resource.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Name: faker.Word(),
	}

	_, err := st.Store(ctx, meta)
	assert.NoError(t, err)

	res, err := fn(ctx, []*resource.Meta{meta})
	assert.NoError(t, err)
	assert.Len(t, res, 1)
}

func TestUpdateResource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	st := resource.NewStore[*resource.Meta]()
	fn := UpdateResource(st)

	meta := &resource.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Name: faker.Word(),
	}

	_, err := st.Store(ctx, meta)
	assert.NoError(t, err)

	res, err := fn(ctx, []*resource.Meta{meta})
	assert.NoError(t, err)
	assert.Len(t, res, 1)
}

func TestDeleteResource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	st := resource.NewStore[*resource.Meta]()
	fn := DeleteResource(st)

	meta := &resource.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Name: faker.Word(),
	}

	_, err := st.Store(ctx, meta)
	assert.NoError(t, err)

	res, err := fn(ctx, []*resource.Meta{meta})
	assert.NoError(t, err)
	assert.Len(t, res, 1)
}
