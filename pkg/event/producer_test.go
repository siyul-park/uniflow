package event

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestProducer_Produce(t *testing.T) {
	topic := faker.Word()

	q := NewQueue(0)
	defer q.Close()
	p := NewProducer(q)

	e := New(topic)

	p.Produce(e)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case r := <-q.Pop():
		assert.Equal(t, e, r)
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}
