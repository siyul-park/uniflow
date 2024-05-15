package event

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	d := faker.UUIDHyphenated()

	e := New(d)
	defer e.Close()

	assert.Equal(t, d, e.Data())
}

func TestEvent_Close(t *testing.T) {
	e := New(nil)
	e.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case <-e.Done():
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}
