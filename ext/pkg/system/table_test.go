package system

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestTable_LoadAndStore(t *testing.T) {
	opcode := faker.Word()

	tb := NewTable()
	tb.Store(opcode, func() {})

	_, err := tb.Load(opcode)
	assert.NoError(t, err)
}
