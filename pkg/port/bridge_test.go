package port

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/stretchr/testify/assert"
)

func TestBridge_WriteAndReceive(t *testing.T) {
	t.Run("NoOutput", func(t *testing.T) {
		b := NewBridge()
		defer b.Close()

		w := NewWriter()
		defer w.Close()

		r := NewReader()
		defer r.Close()

		w.Link(r)

		pck1 := packet.New(nil)

		w.Write(pck1)
		<-r.Read()

		count := b.Write(nil, []*Reader{r}, nil)
		assert.Equal(t, 0, count)

		pck2 := <-w.Receive()
		assert.Equal(t, packet.None, pck2)
	})

	t.Run("SingleOutput", func(t *testing.T) {
		b := NewBridge()
		defer b.Close()

		w1 := NewWriter()
		defer w1.Close()

		w2 := NewWriter()
		defer w2.Close()

		r1 := NewReader()
		defer r1.Close()

		r2 := NewReader()
		defer r2.Close()

		w1.Link(r1)
		w2.Link(r2)

		pck1 := packet.New(nil)

		w1.Write(pck1)
		<-r1.Read()

		count := b.Write([]*packet.Packet{pck1}, []*Reader{r1}, []*Writer{w2})
		assert.Equal(t, 1, count)

		pck2 := <-r2.Read()
		assert.Equal(t, pck1, pck2)

		r2.Receive(pck2)
		<-w2.Receive()

		ok := b.Receive(pck2, w2)
		assert.True(t, ok)

		pck3 := <-w1.Receive()
		assert.Equal(t, pck1, pck3)
	})

	t.Run("MultipleOutputs", func(t *testing.T) {
		b := NewBridge()
		defer b.Close()

		w1 := NewWriter()
		defer w1.Close()

		w2 := NewWriter()
		defer w2.Close()

		w3 := NewWriter()
		defer w3.Close()

		r1 := NewReader()
		defer r1.Close()

		r2 := NewReader()
		defer r2.Close()

		r3 := NewReader()
		defer r3.Close()

		w1.Link(r1)
		w2.Link(r2)
		w3.Link(r3)

		pck1 := packet.New(nil)

		w1.Write(pck1)
		<-r1.Read()

		count := b.Write([]*packet.Packet{pck1, pck1}, []*Reader{r1}, []*Writer{w2, w3})
		assert.Equal(t, 2, count)

		pck2 := <-r2.Read()
		assert.Equal(t, pck1, pck2)

		r2.Receive(pck2)
		<-w2.Receive()

		ok := b.Receive(pck2, w2)
		assert.True(t, ok)

		pck3 := <-r3.Read()
		assert.Equal(t, pck1, pck3)

		r3.Receive(pck3)
		<-w3.Receive()

		ok = b.Receive(pck3, w3)
		assert.True(t, ok)

		pck4 := <-w1.Receive()
		assert.NotNil(t, pck4)
	})
}

func BenchmarkBridge_WriteAndReceive(b *testing.B) {
	br := NewBridge()
	defer br.Close()

	w1 := NewWriter()
	defer w1.Close()

	w2 := NewWriter()
	defer w2.Close()

	r1 := NewReader()
	defer r1.Close()

	r2 := NewReader()
	defer r2.Close()

	w1.Link(r1)
	w2.Link(r2)

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			pck1 := packet.New(nil)

			w1.Write(pck1)
			<-r1.Read()

			count := br.Write([]*packet.Packet{pck1}, []*Reader{r1}, []*Writer{w2})
			assert.Equal(b, 1, count)

			pck2 := <-r2.Read()
			assert.Equal(b, pck1, pck2)

			r2.Receive(pck2)
			<-w2.Receive()

			ok := br.Receive(pck2, w2)
			assert.True(b, ok)

			pck3 := <-w1.Receive()
			assert.Equal(b, pck1, pck3)
		}
	})
}
