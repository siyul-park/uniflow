package port

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/stretchr/testify/assert"
)

func TestBridge_WriteAndReceive(t *testing.T) {
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
}
