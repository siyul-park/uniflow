package printer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTablePrinter_Print(t *testing.T) {
	data := []map[string]any{
		{
			"foo": "foo",
			"bar": "bar",
		},
	}

	p, err := NewTable([]TableColumnDefinition{
		{Name: "foo", Format: "$.foo"},
		{Name: "bar", Format: "$.bar"},
	})
	assert.NoError(t, err)

	expect := " FOO  BAR \n foo  bar "

	table, err := p.Print(data)
	assert.NoError(t, err)
	assert.Equal(t, expect, table)
}
