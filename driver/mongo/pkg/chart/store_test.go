package chart

import (
	"context"
	"github.com/siyul-park/uniflow/pkg/spec"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/driver/mongo/pkg/server"
	"github.com/siyul-park/uniflow/pkg/chart"
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

	chrt := &chart.Chart{
		ID: uuid.Must(uuid.NewV7()),
	}

	_, _ = st.Store(ctx, chrt)
	_, _ = st.Store(ctx, chrt)
	_, _ = st.Delete(ctx, chrt)
}

func TestStore_Load(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := server.New()
	defer server.Release(s)

	c, _ := mongo.Connect(options.Client().ApplyURI(s.URI()))
	defer c.Disconnect(ctx)

	st := NewStore(c.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

	chrt1 := &chart.Chart{
		ID: uuid.Must(uuid.NewV7()),
	}
	chrt2 := &chart.Chart{
		ID: uuid.Must(uuid.NewV7()),
	}

	count, err := st.Store(ctx, chrt1, chrt2)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)

	loaded, err := st.Load(ctx, chrt1, chrt2)
	assert.NoError(t, err)
	assert.Len(t, loaded, 2)
}

func TestStore_Store(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := server.New()
	defer server.Release(s)

	c, _ := mongo.Connect(options.Client().ApplyURI(s.URI()))
	defer c.Disconnect(ctx)

	st := NewStore(c.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

	chrt1 := &chart.Chart{
		ID: uuid.Must(uuid.NewV7()),
		Specs: []*spec.Unstructured{
			{
				Meta: spec.Meta{
					ID:   uuid.Must(uuid.NewV7()),
					Kind: faker.UUIDHyphenated(),
				},
			},
		},
	}
	chrt2 := &chart.Chart{
		ID: uuid.Must(uuid.NewV7()),
	}

	count, err := st.Store(ctx, chrt1, chrt2)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)

	loaded, err := st.Load(ctx, chrt1, chrt2)
	assert.NoError(t, err)
	assert.Len(t, loaded, 2)
}

func TestStore_Swap(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := server.New()
	defer server.Release(s)

	c, _ := mongo.Connect(options.Client().ApplyURI(s.URI()))
	defer c.Disconnect(ctx)

	st := NewStore(c.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

	chrt1 := &chart.Chart{
		ID: uuid.Must(uuid.NewV7()),
	}
	chrt2 := &chart.Chart{
		ID: uuid.Must(uuid.NewV7()),
	}

	count, err := st.Store(ctx, chrt1, chrt2)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)

	count, err = st.Swap(ctx, chrt1, chrt2)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)

	loaded, err := st.Load(ctx, chrt1, chrt2)
	assert.NoError(t, err)
	assert.Len(t, loaded, 2)
}

func TestMemStore_Delete(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := server.New()
	defer server.Release(s)

	c, _ := mongo.Connect(options.Client().ApplyURI(s.URI()))
	defer c.Disconnect(ctx)

	st := NewStore(c.Database(faker.UUIDHyphenated()).Collection(faker.UUIDHyphenated()))

	chrt1 := &chart.Chart{
		ID: uuid.Must(uuid.NewV7()),
	}
	chrt2 := &chart.Chart{
		ID: uuid.Must(uuid.NewV7()),
	}

	count, err := st.Store(ctx, chrt1, chrt2)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)

	count, err = st.Delete(ctx, chrt1, chrt2)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)

	loaded, err := st.Load(ctx, chrt1, chrt2)
	assert.NoError(t, err)
	assert.Len(t, loaded, 0)
}
