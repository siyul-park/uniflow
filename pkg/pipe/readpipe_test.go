package pipe

import (
	"context"
	"fmt"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestReadPipe_Close(t *testing.T) {
	p := NewRead[string](0)

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
	p := NewRead[string](0)
	defer p.Close()

	data1 := faker.UUIDHyphenated()
	data2 := faker.UUIDHyphenated()

	p.write(data1)
	p.write(data2)

	assert.Equal(t, data1, <-p.Read())
	assert.Equal(t, data2, <-p.Read())
}

func BenchmarkReadPipe_Read(b *testing.B) {
	for i := -1; i < 4; i++ {
		capacity := 0
		if i >= 0 {
			capacity = int(math.Pow(2, float64(i)))
		}

		b.Run(fmt.Sprintf("Capacity: %d", capacity), func(b *testing.B) {
			p := NewRead[string](capacity)
			defer p.Close()

			wg := sync.WaitGroup{}
			for i := 0; i < b.N; i++ {
				data := faker.UUIDHyphenated()
				p.write(data)

				wg.Add(1)
				go func() {
					defer wg.Done()
					<-p.Read()
				}()
			}

			wg.Wait()
		})
	}
}
