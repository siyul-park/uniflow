package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollector_Read(t *testing.T) {
	r1 := NewReader()
	defer r1.Close()

	r2 := NewReader()
	defer r2.Close()

	c := NewCollector([]*Reader{r1, r2})
	defer c.Close()

	pck := New(nil)

	reads := c.Read(r1, pck)
	assert.Len(t, reads, 0)

	reads = c.Read(r2, pck)
	assert.Len(t, reads, 2)
}
