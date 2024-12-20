package spec

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/driver/mongo/pkg/server"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestStore_Index(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := server.New()
	defer server.Release(s)

	c, _ := mongo.Connect(options.Client().ApplyURI(s.URI()))
	defer c.Disconnect(ctx)

	st := NewStore(c.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

	err := st.Index(ctx)
	assert.NoError(t, err)
}

func TestStore_Watch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := server.New()
	defer server.Release(s)

	c, _ := mongo.Connect(options.Client().ApplyURI(s.URI()))
	defer c.Disconnect(ctx)

	st := NewStore(c.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

	stream, err := st.Watch(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, stream)

	defer stream.Close()

	go func() {
		for {
			if event, ok := <-stream.Next(); ok {
				assert.NotZero(t, event.ID)
			} else {
				return
			}
		}
	}()

	meta := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: resource.DefaultNamespace,
	}

	_, _ = st.Store(ctx, meta)
	_, _ = st.Store(ctx, meta)
	_, _ = st.Delete(ctx, meta)
}

func TestStore_Load(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := server.New()
	defer server.Release(s)

	c, _ := mongo.Connect(options.Client().ApplyURI(s.URI()))
	defer c.Disconnect(ctx)

	st := NewStore(c.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

	meta1 := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}
	meta2 := &spec.Meta{
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

	s := server.New()
	defer server.Release(s)

	c, _ := mongo.Connect(options.Client().ApplyURI(s.URI()))
	defer c.Disconnect(ctx)

	st := NewStore(c.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

	meta1 := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}
	meta2 := &spec.Meta{
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

	s := server.New()
	defer server.Release(s)

	c, _ := mongo.Connect(options.Client().ApplyURI(s.URI()))
	defer c.Disconnect(ctx)

	st := NewStore(c.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

	meta1 := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}
	meta2 := &spec.Meta{
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

func TestMemStore_Delete(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := server.New()
	defer server.Release(s)

	c, _ := mongo.Connect(options.Client().ApplyURI(s.URI()))
	defer c.Disconnect(ctx)

	st := NewStore(c.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

	meta1 := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}
	meta2 := &spec.Meta{
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
