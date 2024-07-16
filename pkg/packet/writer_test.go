package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	w := NewWriter()
	defer w.Close()

	r := NewReader()
	defer r.Close()

	w.Link(r)

	go func() {
		for {
			inPck, ok := <-r.Read()
			if !ok {
				return
			}
			r.Receive(inPck)
		}
	}()

	outPck := New(nil)

	backPck := Write(w, outPck)
	assert.Equal(t, outPck, backPck)
}

func TestCallOrReturn(t *testing.T) {
	t.Run("Call", func(t *testing.T) {
		w := NewWriter()
		defer w.Close()

		r := NewReader()
		defer r.Close()

		w.Link(r)

		go func() {
			for {
				inPck, ok := <-r.Read()
				if !ok {
					return
				}
				r.Receive(inPck)
			}
		}()

		outPck := New(nil)

		backPck := WriteOrFallback(w, outPck, None)
		assert.Equal(t, outPck, backPck)
	})

	t.Run("Return", func(t *testing.T) {
		w := NewWriter()
		defer w.Close()

		outPck := New(nil)

		backPck := WriteOrFallback(w, outPck, None)
		assert.Equal(t, None, backPck)
	})
}

func TestWriter_Write(t *testing.T) {
	w := NewWriter()
	defer w.Close()

	r := NewReader()
	defer r.Close()

	w.Link(r)

	out1 := New(nil)
	out2 := New(nil)

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

	b.RunParallel(func(p *testing.PB) {
		out := New(nil)

		for p.Next() {
			count := w.Write(out)
			assert.Equal(b, 1, count)

			in, ok := <-r.Read()
			assert.True(b, ok)
			assert.Equal(b, out, in)
		}
	})
}
