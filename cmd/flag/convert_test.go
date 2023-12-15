package flag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToShorthand(t *testing.T) {
	flag := "camelCase"

	s := ToShorthand(flag)
	assert.Equal(t, flag[0:1], s)
}
