package packet

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/types"

	"github.com/stretchr/testify/assert"
)

func TestTracer_Dispatch(t *testing.T) {
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

	tr := NewTracer()
	defer tr.Close()

	pck1 := New(nil)
	pck2 := New(nil)

	w1.Write(pck1)
	<-r1.Read()

	tr.Read(r1, pck1)
	tr.Write(w2, pck1)

	<-r2.Read()
	r2.Receive(pck2)
	w2.Receive()

	count := 0
	tr.Dispatch(pck1, HookFunc(func(pck *Packet) {
		count++
	}))

	tr.Receive(w2, pck2)
	assert.Equal(t, 1, count)
}

func TestTracer_Link(t *testing.T) {
	tr := NewTracer()
	defer tr.Close()

	pck1 := New(nil)
	pck2 := New(nil)

	tr.Link(pck1, pck2)

	assert.Equal(t, tr.Links(pck1, nil), []*Packet{pck1, pck2})
	assert.Equal(t, tr.Links(nil, pck2), []*Packet{pck2, pck1})
	assert.Equal(t, tr.Links(pck1, pck2), []*Packet{pck1, pck2})
}

func TestTracer_Read(t *testing.T) {
	w := NewWriter()
	defer w.Close()

	r := NewReader()
	defer r.Close()

	w.Link(r)

	tr := NewTracer()
	defer tr.Close()

	pck := New(nil)

	w.Write(pck)
	<-r.Read()

	tr.Read(r, pck)
	assert.Contains(t, tr.Reads(r), pck)
}

func TestTracer_Write(t *testing.T) {
	w := NewWriter()
	defer w.Close()

	r := NewReader()
	defer r.Close()

	w.Link(r)

	tr := NewTracer()
	defer tr.Close()

	pck1 := New(types.NewString(faker.UUIDHyphenated()))

	tr.Write(w, pck1)
	assert.Contains(t, tr.Writes(w), pck1)

	pck2, ok := <-r.Read()
	assert.True(t, ok)
	assert.Equal(t, pck2.Payload(), pck1.Payload())
}

func TestTracer_Receive(t *testing.T) {
	t.Run("Receive", func(t *testing.T) {
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

		tr := NewTracer()
		defer tr.Close()

		pck1 := New(nil)
		pck2 := New(nil)
		pck3 := New(nil)

		w1.Write(pck1)
		<-r1.Read()

		tr.Read(r1, pck1)
		tr.Link(pck1, pck3)
		tr.Write(w2, pck3)

		<-r2.Read()
		r2.Receive(pck2)
		w2.Receive()

		tr.Receive(w2, pck2)

		pck4, ok := <-w1.Receive()
		assert.True(t, ok)
		assert.Equal(t, pck2, pck4)

		assert.Len(t, tr.Writes(w2), 0)
		assert.Len(t, tr.Reads(r1), 0)
		assert.Len(t, tr.Receives(pck3), 0)
		assert.Len(t, tr.Receives(pck1), 0)
	})

	t.Run("Discard", func(t *testing.T) {
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

		tr := NewTracer()
		defer tr.Close()

		pck1 := New(nil)
		pck2 := New(nil)
		pck3 := New(nil)

		w1.Write(pck1)
		<-r1.Read()

		tr.Read(r1, pck1)
		tr.Link(pck1, pck3)
		tr.Write(w2, pck3)

		<-r2.Read()
		r2.Receive(pck2)
		w2.Receive()

		tr.Receive(w2, nil)

		pck4, ok := <-w1.Receive()
		assert.True(t, ok)
		assert.Equal(t, None, pck4)

		assert.Len(t, tr.Writes(w2), 0)
		assert.Len(t, tr.Reads(r1), 0)
		assert.Len(t, tr.Receives(pck3), 0)
		assert.Len(t, tr.Receives(pck1), 0)
	})
}
