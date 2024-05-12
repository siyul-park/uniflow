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
	assert.Equal(t, d, e.Data())
}

func TestEvent_Ref(t *testing.T) {
	e := New(nil)
	e.Wait(1)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		<-e.Done()
		close(done)
	}()

	e.Wait(-1)

	select {
	case <-done:
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}
