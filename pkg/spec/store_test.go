package spec

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/stretchr/testify/assert"
)

func TestStore_Index(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	st := NewStore(memdb.NewCollection(""))

	err := st.Index(ctx)
	assert.NoError(t, err)

	err = st.Index(ctx)
	assert.NoError(t, err)
}

func TestStore_Watch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	stream, err := st.Watch(ctx, nil)
	assert.NoError(t, err)
	assert.NotNil(t, stream)

	defer stream.Close()

	go func() {
		for {
			if event, ok := <-stream.Next(); ok {
				assert.NotNil(t, event.ID)
			} else {
				return
			}
		}
	}()

	meta := &Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: DefaultNamespace,
	}

	_, _ = st.Store(ctx, meta)
	_, _ = st.Store(ctx, meta)
	_, _ = st.Delete(ctx, meta)
}

func TestStore_Load(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	meta1 := &Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}
	meta2 := &Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	count, err := st.Store(ctx, meta1, meta2)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)

	loaded, err := st.Load(ctx, meta1, meta2)
	assert.NoError(t, err)
	assert.Len(t, loaded, 2)
}

func TestStore_Store(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	meta1 := &Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}
	meta2 := &Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	count, err := st.Store(ctx, meta1, meta2)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)

	loaded, err := st.Load(ctx, meta1, meta2)
	assert.NoError(t, err)
	assert.Len(t, loaded, 2)
}

func TestStore_Swap(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	meta1 := &Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}
	meta2 := &Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	count, err := st.Store(ctx, meta1, meta2)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)

	count, err = st.Swap(ctx, meta1, meta2)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)

	loaded, err := st.Load(ctx, meta1, meta2)
	assert.NoError(t, err)
	assert.Len(t, loaded, 2)
}

func TestStore_Delete(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	st := NewStore(memdb.NewCollection(""))

	meta1 := &Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}
	meta2 := &Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	count, err := st.Store(ctx, meta1, meta2)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)

	count, err = st.Delete(ctx, meta1, meta2)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)

	loaded, err := st.Load(ctx, meta1, meta2)
	assert.NoError(t, err)
	assert.Len(t, loaded, 0)
}
