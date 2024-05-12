package event

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestBroker_ProduceAndConsume(t *testing.T) {
	topic := faker.Word()

	b := NewBroker()
	defer b.Close()

	p := b.Producer()

	c1 := b.Consumer(topic)
	defer c1.Close()

	c2 := b.Consumer(topic)
	defer c2.Close()

	e := New(topic)

	p.Produce(e)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case r := <-c1.Consume():
		assert.Equal(t, e, r)
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}

	select {
	case r := <-c2.Consume():
		assert.Equal(t, e, r)
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}
