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
	case e := <-q.Pop():
		assert.Equal(t, e1, e)
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}

	select {
	case e := <-q.Pop():
		assert.Equal(t, e2, e)
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}

func TestQueue_CloseAndDone(t *testing.T) {
	q := NewQueue(0)
	q.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case <-q.Done():
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}
