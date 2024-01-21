package js

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	vm := New()
	assert.NotNil(t, vm)
}
