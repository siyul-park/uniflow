package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadGroup_Read(t *testing.T) {
	r1 := NewReader()
	defer r1.Close()

	r2 := NewReader()
	defer r2.Close()

	rg := NewReadGroup([]*Reader{r1, r2})
	defer rg.Close()

	pck := New(nil)

	reads := rg.Read(r1, pck)
	assert.Len(t, reads, 0)

	reads = rg.Read(r2, pck)
	assert.Len(t, reads, 2)
}
