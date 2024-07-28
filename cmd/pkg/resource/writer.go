package resource

import (
	"encoding/json"
	"io"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// Writer writes structured data to an io.Writer in table format.
type Writer struct {
	writer io.Writer
}

var style = table.Style{
	Name: "StyleDefault",
	Box: table.BoxStyle{
		EmptySeparator: text.RepeatAndTrim(" ", text.RuneWidthWithoutEscSequences(" ")),
		PaddingLeft:    " ",
		PaddingRight:   " ",
		PageSeparator:  "\n",
		UnfinishedRow:  " ~",
	},
	Color:  table.ColorOptionsDefault,
	Format: table.FormatOptionsDefault,
	HTML:   table.DefaultHTMLOptions,
	Options: table.Options{
		DoNotColorBordersAndSeparators: true,
	},
	Title: table.TitleOptionsDefault,
}

// NewWriter creates a new Writer for the given io.Writer.
func NewWriter(writer io.Writer) *Writer {
	return &Writer{writer: writer}
}

// Write encodes the value, transforms it into a table, and writes it to the writer.
func (w *Writer) Write(value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	var elements []map[string]any
	if err := json.Unmarshal(data, &elements); err != nil {
		var element map[string]any
		if err := json.Unmarshal(data, &element); err != nil {
			return err
		}
		elements = append(elements, element)
	}

	matrix := map[string]int{}
	for _, element := range elements {
		for key := range element {
			matrix[key]++
		}
	}

	var keys []string
	for key, count := range matrix {
		if count > len(elements)/3 {
			keys = append(keys, key)
		}
	}

	header := table.Row{}
	for _, key := range keys {
		header = append(header, key)
	}

	rows := make([]table.Row, 0, len(elements))
	for _, element := range elements {
		row := make(table.Row, 0, len(header))
		for _, key := range keys {
			row = append(row, element[key])
		}
		rows = append(rows, row)
	}

	tb := table.NewWriter()
	tb.SetStyle(style)
	tb.AppendHeader(header)
	tb.AppendRows(rows)

	_, err = w.writer.Write([]byte(tb.Render()))
	return err
}
