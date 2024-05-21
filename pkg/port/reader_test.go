package port

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/stretchr/testify/assert"
)

func TestReader_Receive(t *testing.T) {
	w := NewWriter()
	defer w.Close()

	r := NewReader()
	defer r.Close()

	w.Link(r)

	out1 := packet.New(nil)
	out2 := packet.New(nil)

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

	out := packet.New(nil)
	for i := 0; i < b.N; i++ {
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
}
