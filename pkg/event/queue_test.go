package event

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueue_PushAndPop(t *testing.T) {
	q := NewQueue(0)
	defer q.Close()

	e1 := New(nil)
	e2 := New(nil)

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
