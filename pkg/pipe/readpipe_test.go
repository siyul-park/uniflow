package pipe

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestReadPipe_Close(t *testing.T) {
	p := newRead[string](0)

	p.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case <-p.Done():
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}

func TestReadPipe_Read(t *testing.T) {
	t.Run("Not Closed", func(t *testing.T) {
		p := newRead[string](0)
		defer p.Close()

		data1 := faker.UUIDHyphenated()
		data2 := faker.UUIDHyphenated()

		p.write(data1)
		p.write(data2)

		assert.Equal(t, data1, <-p.Read())
		assert.Equal(t, data2, <-p.Read())
	})

	t.Run("Closed", func(t *testing.T) {
		p := newRead[string](0)
		p.Close()

		data1 := faker.UUIDHyphenated()
		data2 := faker.UUIDHyphenated()

		p.write(data1)
		p.write(data2)

		assert.Zero(t, <-p.Read())
		assert.Zero(t, <-p.Read())
	})
}
