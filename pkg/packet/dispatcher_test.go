package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDispatcher_Read(t *testing.T) {
	r1 := NewReader()
	defer r1.Close()

	r2 := NewReader()
	defer r2.Close()

	d := NewDispatcher([]*Reader{r1, r2})
	defer d.Close()

	pck := New(nil)

	reads := d.Read(r1, pck)
	assert.Len(t, reads, 0)

	reads = d.Read(r2, pck)
	assert.Len(t, reads, 2)
}
