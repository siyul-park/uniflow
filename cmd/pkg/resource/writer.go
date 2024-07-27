package resource

import (
	"io"
	"slices"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/siyul-park/uniflow/pkg/types"
)

type Writer struct {
	writer io.Writer
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

func NewWriter(writer io.Writer) *Writer {
	return &Writer{writer: writer}
}

func (w *Writer) Write(value any) error {
	doc, err := types.Encoder.Encode(value)
	if err != nil {
		return err
	}

	var elements []types.Map
	switch v := doc.(type) {
	case types.Slice:
		for _, value := range v.Values() {
			if v, ok := value.(types.Map); ok {
				elements = append(elements, v)
			}
		}
	case types.Map:
		elements = append(elements, v)
	}

	metrix := map[string]int{}
	for _, element := range elements {
		for _, key := range element.Keys() {
			if k, ok := key.(types.String); ok {
				metrix[k.String()]++
			}
		}
	}

	var keys []string
	for key, count := range metrix {
		if count > len(elements)/2 {
			keys = append(keys, key)
		}
	}

	slices.Sort(keys)

	header := table.Row{}
	for _, key := range keys {
		header = append(header, key)
	}

	rows := make([]table.Row, 0, len(elements))
	for _, element := range elements {
		row := make(table.Row, 0, len(header))
		for _, key := range keys {
			val, _ := element.Get(types.NewString(key))
			row = append(row, types.InterfaceOf(val))
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
