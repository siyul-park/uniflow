package event

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConsumer_Consume(t *testing.T) {
	q := NewQueue(0)
	defer q.Close()

	c := NewConsumer(q)
	e := New(nil)

	q.Push(e)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case r := <-c.Consume():
		assert.Equal(t, e, r)
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}
