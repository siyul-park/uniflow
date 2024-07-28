package secret

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestStore_Watch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	st := NewStore()

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

	secret := &Secret{
		ID:        uuid.Must(uuid.NewV7()),
		Namespace: DefaultNamespace,
	}

	_, _ = st.Store(ctx, secret)
	_, _ = st.Store(ctx, secret)
	_, _ = st.Delete(ctx, secret)
}

func TestStore_Load(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	st := NewStore()

	secret1 := &Secret{
		ID: uuid.Must(uuid.NewV7()),
	}
	secret2 := &Secret{
		ID: uuid.Must(uuid.NewV7()),
	}

	count, err := st.Store(ctx, secret1, secret2)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)

	loaded, err := st.Load(ctx, secret1, secret2)
	assert.NoError(t, err)
	assert.Len(t, loaded, 2)
}

func TestStore_Store(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	st := NewStore()

	secret1 := &Secret{
		ID: uuid.Must(uuid.NewV7()),
	}
	secret2 := &Secret{
		ID: uuid.Must(uuid.NewV7()),
	}

	count, err := st.Store(ctx, secret1, secret2)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)

	loaded, err := st.Load(ctx, secret1, secret2)
	assert.NoError(t, err)
	assert.Len(t, loaded, 2)
}

func TestStore_Swap(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	st := NewStore()

	secret1 := &Secret{
		ID: uuid.Must(uuid.NewV7()),
	}
	secret2 := &Secret{
		ID: uuid.Must(uuid.NewV7()),
	}

	count, err := st.Store(ctx, secret1, secret2)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)

	count, err = st.Swap(ctx, secret1, secret2)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)

	loaded, err := st.Load(ctx, secret1, secret2)
	assert.NoError(t, err)
	assert.Len(t, loaded, 2)
}

func TestStore_Delete(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	st := NewStore()

	secret1 := &Secret{
		ID: uuid.Must(uuid.NewV7()),
	}
	secret2 := &Secret{
		ID: uuid.Must(uuid.NewV7()),
	}

	count, err := st.Store(ctx, secret1, secret2)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)

	count, err = st.Delete(ctx, secret1, secret2)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)

	loaded, err := st.Load(ctx, secret1, secret2)
	assert.NoError(t, err)
	assert.Len(t, loaded, 0)
}
