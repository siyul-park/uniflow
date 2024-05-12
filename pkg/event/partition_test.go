package event

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPartition_WriteAndRead(t *testing.T) {
	p := NewPartition()
	defer p.Close()

	p1 := p.Producer()

	c1 := p.Consumer()
	defer c1.Close()

	c2 := p.Consumer()
	defer c2.Close()

	e := New()

	p1.Produce(e)

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
