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
	v := faker.UUIDHyphenated()

	assert.True(t, h.Store(k, v))
	assert.False(t, h.Store(k, v))
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
