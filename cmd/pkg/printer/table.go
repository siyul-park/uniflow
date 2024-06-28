package printer

import (
	"errors"
	"fmt"
	"io"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/oliveagle/jsonpath"
	"github.com/siyul-park/uniflow/pkg/object"
)

// TableColumnDefinition represents the definition of a table column.
type TableColumnDefinition struct {
	Name   string
	Format string
}

// TablePrinter is responsible for printing tabular data based on the provided columns.
type TablePrinter struct {
	names   []string
	formats []string
}

// SpecTableColumnDefinitions defines columns for displaying spec information.
var SpecTableColumnDefinitions = []TableColumnDefinition{
	{Name: "id", Format: "$.id"},
	{Name: "kind", Format: "$.kind"},
	{Name: "namespace", Format: "$.namespace"},
	{Name: "name", Format: "$.name"},
	{Name: "links", Format: "$.links"},
}

var style = table.Style{
	Name: "StyleDefault",
	Box: table.BoxStyle{
		BottomLeft:       "",
		BottomRight:      "",
		BottomSeparator:  "",
		EmptySeparator:   text.RepeatAndTrim(" ", text.RuneWidthWithoutEscSequences(" ")),
		Left:             "",
		LeftSeparator:    "",
		MiddleHorizontal: "",
		MiddleSeparator:  "",
		MiddleVertical:   "",
		PaddingLeft:      " ",
		PaddingRight:     " ",
		PageSeparator:    "\n",
		Right:            "",
		RightSeparator:   "",
		TopLeft:          "",
		TopRight:         "",
		TopSeparator:     "",
		UnfinishedRow:    " ~",
	},
	Color:  table.ColorOptionsDefault,
	Format: table.FormatOptionsDefault,
	HTML:   table.DefaultHTMLOptions,
	Options: table.Options{
		DoNotColorBordersAndSeparators: true,
	},
	Title: table.TitleOptionsDefault,
}

// PrintTable prints tabular data to the specified writer using the provided columns.
func PrintTable(writer io.Writer, data any, columns []TableColumnDefinition) error {
	tablePrinter, err := NewTable(columns)
	if err != nil {
		return err
	}

	table, err := tablePrinter.Print(data)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprint(writer, table); err != nil {
		return err
	}

	return nil
}

// NewTable creates a new TablePrinter based on the provided column definitions.
func NewTable(columns []TableColumnDefinition) (*TablePrinter, error) {
	names := make([]string, len(columns))
	formats := make([]string, len(columns))

	for i, column := range columns {
		names[i] = column.Name
		formats[i] = column.Format
	}

	return &TablePrinter{
		names:   names,
		formats: formats,
	}, nil
}

// Print formats and prints the provided data as a table.
func (p *TablePrinter) Print(data any) (string, error) {
	value, err := object.MarshalText(data)
	if err != nil {
		return "", err
	}

	var elements []any
	switch v := value.(type) {
	case object.Slice:
		elements = v.Slice()
	case object.Map:
		elements = append(elements, v.Interface())
	default:
		return "", errors.New("unsupported data type")
	}

	header := make(table.Row, len(p.names))
	for i, name := range p.names {
		header[i] = name
	}

	rows := make([]table.Row, len(elements))
	for i, element := range elements {
		row := make(table.Row, len(p.formats))
		for j, format := range p.formats {
			data, _ := jsonpath.JsonPathLookup(element, format)
			row[j] = data
		}
		rows[i] = row
	}

	tb := table.NewWriter()
	tb.SetStyle(style)
	tb.AppendHeader(header)
	tb.AppendRows(rows)

	return tb.Render(), nil
}
