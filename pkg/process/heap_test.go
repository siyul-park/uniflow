package process

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestHeap_Load(t *testing.T) {
	h := newHeap()
	defer h.Close()

	k := faker.UUIDHyphenated()
	v := faker.UUIDHyphenated()

	r := h.Load(k)
	assert.Equal(t, nil, r)

	h.Store(k, v)

	r = h.Load(k)
	assert.Equal(t, v, r)
}

func TestHeap_Store(t *testing.T) {
	h := newHeap()
	defer h.Close()

	k := faker.UUIDHyphenated()
	v1 := faker.UUIDHyphenated()
	v2 := faker.UUIDHyphenated()

	h.Store(k, v1)
	h.Store(k, v2)

	r := h.Load(k)
	assert.Equal(t, v2, r)
}

func TestHeap_Delete(t *testing.T) {
	h := newHeap()
	defer h.Close()

	k := faker.UUIDHyphenated()
	v := faker.UUIDHyphenated()

	ok := h.Delete(k)
	assert.False(t, ok)

	h.Store(k, v)

	ok = h.Delete(k)
	assert.True(t, ok)
}
