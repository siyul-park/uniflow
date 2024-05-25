package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDispatcher_Write(t *testing.T) {
	r := NewReader()
	defer r.Close()

	d := NewDispatcher([]*Reader{r}, RouteHookFunc(func(pcks []*Packet) bool {
		return true
	}))
	defer d.Close()

	pck := New(nil)

	count := d.Write(pck, r)
	assert.Equal(t, 1, count)
}

func TestDispatcher_Forward(t *testing.T) {
	t.Run("Forward", func(t *testing.T) {
		r := NewReader()
		defer r.Close()

		count := 0
		d := NewDispatcher([]*Reader{r}, RouteHookFunc(func(pcks []*Packet) bool {
			count++
			return true
		}))
		defer d.Close()

		pck := New(nil)

		d.Write(pck, r)
		assert.Equal(t, 1, count)

		d.Write(pck, r)
		assert.Equal(t, 2, count)
	})

	t.Run("Drop", func(t *testing.T) {
		w := NewWriter()
		defer w.Close()

		r := NewReader()
		defer r.Close()

		w.Link(r)

		count := 0
		d := NewDispatcher([]*Reader{r}, RouteHookFunc(func(pcks []*Packet) bool {
			count++
			return false
		}))
		defer d.Close()

		pck := New(nil)

		w.Write(pck)
		w.Write(pck)

		<-r.Read()
		<-r.Read()

		d.Write(pck, r)
		assert.Equal(t, 1, count)
		assert.Equal(t, None, <-w.Receive())

		d.Write(pck, r)
		assert.Equal(t, 2, count)
		assert.Equal(t, None, <-w.Receive())
	})
}
