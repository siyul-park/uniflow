package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	pck := New(nil)
	assert.NotNil(t, pck)
}
