package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTracer_Transform(t *testing.T) {
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

	pck1 := New(nil)
	pck2 := New(nil)
	pck3 := New(nil)

	w1.Write(pck1)
	<-r1.Read()

	tr.Read(r1, pck1)
	tr.Transform(pck1, pck2)
	tr.Write(w2, pck2)

	<-r2.Read()
	r2.Receive(pck3)
	w2.Receive()

	tr.Receive(w2, pck3)

	pck4, ok := <-w1.Receive()
	assert.True(t, ok)
	assert.Equal(t, pck3, pck4)
}

func TestTracer_Redirect(t *testing.T) {
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

	tr := NewTracer()

	pck1 := New(nil)
	pck2 := New(nil)

	w1.Write(pck1)
	<-r1.Read()

	tr.Read(r1, pck1)
	tr.Write(w2, pck1)

	<-r2.Read()
	r2.Receive(pck2)
	w2.Receive()

	w3.Write(pck2)
	tr.Redirect(w2, w3)

	<-r3.Read()
	r3.Receive(pck2)
	w3.Receive()

	tr.Receive(w3, pck2)

	pck3, ok := <-w1.Receive()
	assert.True(t, ok)
	assert.Equal(t, pck2, pck3)
}

func TestTracer_ReadAndWriteAndReceive(t *testing.T) {
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

	pck1 := New(nil)
	pck2 := New(nil)

	w1.Write(pck1)
	<-r1.Read()

	tr.Read(r1, pck1)
	tr.Write(w2, pck1)

	<-r2.Read()
	r2.Receive(pck2)
	w2.Receive()

	tr.Receive(w2, pck2)

	pck3, ok := <-w1.Receive()
	assert.True(t, ok)
	assert.Equal(t, pck2, pck3)
}
