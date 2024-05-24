package port

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/stretchr/testify/assert"
)

func TestWriter_Write(t *testing.T) {
	w := NewWriter()
	defer w.Close()

	r := NewReader()
	defer r.Close()

	w.Link(r)

	out1 := packet.New(nil)
	out2 := packet.New(nil)

	count := w.Write(out1)
	assert.Equal(t, 1, count)

	count = w.Write(out2)
	assert.Equal(t, 1, count)

	in1, ok := <-r.Read()
	assert.True(t, ok)
	assert.Equal(t, out1, in1)

	in2, ok := <-r.Read()
	assert.True(t, ok)
	assert.Equal(t, out2, in2)
}

func BenchmarkWriter_Write(b *testing.B) {
	w := NewWriter()
	defer w.Close()

	r := NewReader()
	defer r.Close()

	w.Link(r)

	out := packet.New(nil)
	for i := 0; i < b.N; i++ {
		count := w.Write(out)
		assert.Equal(b, 1, count)

		in, ok := <-r.Read()
		assert.True(b, ok)
		assert.Equal(b, out, in)
	}
}
