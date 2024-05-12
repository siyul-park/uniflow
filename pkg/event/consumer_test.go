package event

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestConsumer_Consume(t *testing.T) {
	topic := faker.Word()

	q := NewQueue(0)
	defer q.Close()
	c := NewConsumer(q)
	defer c.Close()

	e := New(topic)

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
