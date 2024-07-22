package printer

import (
	"bytes"
	"testing"

	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestTablePrintTable(t *testing.T) {
	specs := []spec.Spec{
		&spec.Meta{
			Kind: "foo",
			Name: "bar",
		},
	}

	buf := new(bytes.Buffer)

	expect := " ID     KIND  NAMESPACE  NAME  PORTS \n <nil>  foo   <nil>      bar   <nil> "

	err := PrintTable(buf, specs, SpecTableColumnDefinitions)
	assert.NoError(t, err)
	assert.Equal(t, expect, buf.String())
}

func TestTablePrinter_Print(t *testing.T) {
	specs := []spec.Spec{
		&spec.Meta{
			Kind: "foo",
			Name: "bar",
		},
	}

	p, err := NewTable(SpecTableColumnDefinitions)
	assert.NoError(t, err)

	expect := " ID     KIND  NAMESPACE  NAME  PORTS \n <nil>  foo   <nil>      bar   <nil> "

	table, err := p.Print(specs)
	assert.NoError(t, err)
	assert.Equal(t, expect, table)
}
