package event

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestPartition_WriteAndRead(t *testing.T) {
	topic := faker.Word()

	p := NewPartition()
	defer p.Close()
	
	c1 := p.Consumer()
	defer c1.Close()

	c2 := p.Consumer()
	defer c2.Close()

	e := New(topic)

	p.Write(e)

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
