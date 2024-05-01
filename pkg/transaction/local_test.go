package transaction

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestLocal_Load(t *testing.T) {
	l := NewLocal[string]()
	defer l.Close()

	tx := New()
	defer tx.Rollback()

	v := faker.UUIDHyphenated()

	_, ok := l.Load(tx)
	assert.False(t, ok)

	l.Store(tx, v)

	r, ok := l.Load(tx)
	assert.True(t, ok)
	assert.Equal(t, v, r)
}

func TestLocal_Store(t *testing.T) {
	l := NewLocal[string]()
	defer l.Close()

	tx := New()
	defer tx.Rollback()

	v1 := faker.UUIDHyphenated()
	v2 := faker.UUIDHyphenated()

	l.Store(tx, v1)
	l.Store(tx, v2)

	r, ok := l.Load(tx)
	assert.True(t, ok)
	assert.Equal(t, v2, r)
}

func TestLocal_Delete(t *testing.T) {
	l := NewLocal[string]()
	defer l.Close()

	tx := New()
	defer tx.Rollback()

	v := faker.UUIDHyphenated()

	ok := l.Delete(tx)
	assert.False(t, ok)

	l.Store(tx, v)

	ok = l.Delete(tx)
	assert.True(t, ok)
}

func TestLocal_LoadOrStore(t *testing.T) {
	l := NewLocal[string]()
	defer l.Close()

	tx := New()
	defer tx.Rollback()

	v := faker.UUIDHyphenated()

	r, err := l.LoadOrStore(tx, func() (string, error) {
		return v, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, v, r)
}
