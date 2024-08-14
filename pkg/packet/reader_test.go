package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReader_AddHook(t *testing.T) {
	w := NewWriter()
	defer w.Close()

	r := NewReader()
	defer r.Close()

	count := 0
	r.AddInboundHook(HookFunc(func(_ *Packet) {
		count += 1
	}))
	r.AddOutboundHook(HookFunc(func(_ *Packet) {
		count += 1
	}))

	w.Link(r)

	out := New(nil)

	w.Write(out)

	in := <-r.Read()
	assert.Equal(t, 1, count)

	r.Receive(in)
	assert.Equal(t, 2, count)

	back, ok := <-w.Receive()
	assert.True(t, ok)
	assert.Equal(t, in, back)
}

func TestReader_Receive(t *testing.T) {
	w := NewWriter()
	defer w.Close()

	r := NewReader()
	defer r.Close()

	w.Link(r)

	out1 := New(nil)
	out2 := New(nil)

	w.Write(out1)
	w.Write(out2)

	in1 := <-r.Read()
	in2 := <-r.Read()

	ok := r.Receive(in1)
	assert.True(t, ok)

	ok = r.Receive(in2)
	assert.True(t, ok)

	back1, ok := <-w.Receive()
	assert.True(t, ok)
	assert.Equal(t, in1, back1)

	back2, ok := <-w.Receive()
	assert.True(t, ok)
	assert.Equal(t, in2, back2)
}

func BenchmarkReader_Receive(b *testing.B) {
	w := NewWriter()
	defer w.Close()

	r := NewReader()
	defer r.Close()

	w.Link(r)

	b.RunParallel(func(p *testing.PB) {
		out := New(nil)

		for p.Next() {
			b.StopTimer()

			count := w.Write(out)
			assert.Equal(b, 1, count)

			in, ok := <-r.Read()
			assert.True(b, ok)
			assert.Equal(b, out, in)

			b.StartTimer()

			ok = r.Receive(in)
			assert.True(b, ok)

			back, ok := <-w.Receive()
			assert.True(b, ok)
			assert.Equal(b, in, back)
		}
	})
}
