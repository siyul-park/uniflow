package fmt

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"slices"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// Writer writes structured data to an fmt.Writer in table format.
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

// NewWriter creates a new Writer for the given fmt.Writer.
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

	counts := map[string]int{}
	for _, element := range elements {
		for key := range element {
			counts[key]++
		}
	}

	scores := map[string]float64{}
	for _, element := range elements {
		for key, val := range element {
			scores[key] += 1.0 / float64(len(fmt.Sprint(val)))
		}
	}

	for key, score := range scores {
		scores[key] = score/float64(counts[key]) + 1.0/float64(len(key))
	}

	var keys []string
	for key, count := range counts {
		if count >= len(elements)/2 {
			keys = append(keys, key)
		}
	}

	slices.SortFunc(keys, func(x, y string) int {
		if diff := counts[y] - counts[x]; diff != 0 {
			return diff
		}
		if diff := scores[y] - scores[x]; math.Abs(diff) > 0.1 {
			if diff > 0 {
				return 1
			}
			return -1
		}
		return strings.Compare(x, y)
	})

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
