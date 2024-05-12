package event

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestQueue_PushAndPop(t *testing.T) {
	q := NewQueue(0)
	defer q.Close()

	topic := faker.Word()

	e1 := New(topic)
	e2 := New(topic)

	q.Push(e1)
	q.Push(e2)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case r1 := <-q.Pop():
		assert.Equal(t, e1, r1)
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}

	select {
	case r2 := <-q.Pop():
		assert.Equal(t, e2, r2)
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}
