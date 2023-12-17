package printer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTablePrintTable(t *testing.T) {
	data := []map[string]any{
		{
			"foo": "foo",
			"bar": "bar",
		},
	}

	buf := new(bytes.Buffer)

	expect := " FOO  BAR \n foo  bar "

	err := PrintTable(buf, data, []TableColumnDefinition{
		{Name: "foo", Format: "$.foo"},
		{Name: "bar", Format: "$.bar"},
	})
	assert.NoError(t, err)
	assert.Equal(t, expect, buf.String())
}

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
