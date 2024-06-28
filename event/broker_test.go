package event

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestBroker_ProduceAndConsume(t *testing.T) {
	b := NewBroker()
	defer b.Close()

	topic := faker.Word()

	p := b.Producer(topic)
	defer p.Close()

	c := b.Consumer(topic)
	defer c.Close()

	d := faker.UUIDHyphenated()
	e := New(d)

	p.Produce(e)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case r := <-c.Consume():
		r.Close()
		assert.Equal(t, e.Data(), r.Data())
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}

	select {
	case <-e.Done():
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}
