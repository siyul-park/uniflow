package packet

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/types"

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

	outPck := New(types.NewString(faker.UUIDHyphenated()))

	backPck := Send(w, outPck)
	assert.Equal(t, outPck.Payload(), backPck.Payload())
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

		outPck := New(types.NewString(faker.UUIDHyphenated()))

		backPck := SendOrFallback(w, outPck, None)
		assert.Equal(t, outPck.Payload(), backPck.Payload())
	})

	t.Run("Return", func(t *testing.T) {
		w := NewWriter()
		defer w.Close()

		outPck := New(nil)

		backPck := SendOrFallback(w, outPck, None)
		assert.Equal(t, None, backPck)
	})
}

func TestWriter_AddHook(t *testing.T) {
	w := NewWriter()
	defer w.Close()

	r := NewReader()
	defer r.Close()

	count := 0
	w.AddInboundHook(HookFunc(func(_ *Packet) {
		count += 1
	}))
	w.AddOutboundHook(HookFunc(func(_ *Packet) {
		count += 1
	}))

	w.Link(r)

	out := New(nil)

	w.Write(out)
	assert.Equal(t, 1, count)

	in := <-r.Read()

	r.Receive(in)

	back, ok := <-w.Receive()
	assert.True(t, ok)
	assert.Equal(t, in, back)
	assert.Equal(t, 2, count)
}

func TestWriter_Write(t *testing.T) {
	w := NewWriter()
	defer w.Close()

	r := NewReader()
	defer r.Close()

	w.Link(r)

	out1 := New(types.NewString(faker.UUIDHyphenated()))
	out2 := New(types.NewString(faker.UUIDHyphenated()))

	count := w.Write(out1)
	assert.Equal(t, 1, count)

	count = w.Write(out2)
	assert.Equal(t, 1, count)

	in1, ok := <-r.Read()
	assert.True(t, ok)
	assert.Equal(t, out1.Payload(), in1.Payload())

	in2, ok := <-r.Read()
	assert.True(t, ok)
	assert.Equal(t, out2.Payload(), in2.Payload())
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
