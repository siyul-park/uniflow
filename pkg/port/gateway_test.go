package port

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/stretchr/testify/assert"
)

func TestGateway_Write(t *testing.T) {
	r := NewReader()
	defer r.Close()

	g := NewGateway([]*Reader{r}, ForwardHookFunc(func(pcks []*packet.Packet) bool {
		return true
	}))
	defer g.Close()

	pck := packet.New(nil)

	count := g.Write(pck, r)
	assert.Equal(t, 1, count)
}

func TestGateway_Forward(t *testing.T) {
	t.Run("forward", func(t *testing.T) {
		r := NewReader()
		defer r.Close()

		count := 0
		g := NewGateway([]*Reader{r}, ForwardHookFunc(func(pcks []*packet.Packet) bool {
			count++
			return true
		}))
		defer g.Close()

		pck := packet.New(nil)

		g.Write(pck, r)
		assert.Equal(t, 1, count)

		g.Write(pck, r)
		assert.Equal(t, 2, count)
	})

	t.Run("drop", func(t *testing.T) {
		w := NewWriter()
		defer w.Close()

		r := NewReader()
		defer r.Close()

		w.Link(r)

		count := 0
		g := NewGateway([]*Reader{r}, ForwardHookFunc(func(pcks []*packet.Packet) bool {
			count++
			return false
		}))
		defer g.Close()

		pck := packet.New(nil)

		w.Write(pck)
		w.Write(pck)

		<-r.Read()
		<-r.Read()

		g.Write(pck, r)
		assert.Equal(t, 1, count)
		assert.Equal(t, packet.None, <-w.Receive())

		g.Write(pck, r)
		assert.Equal(t, 2, count)
		assert.Equal(t, packet.None, <-w.Receive())
	})
}
