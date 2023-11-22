package printer

import (
	"github.com/iancoleman/strcase"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/xiatechs/jsonata-go"
)

type (
	TableColumnDefinition struct {
		Name   string
		Format string
	}

	TablePrinter struct {
		names   []string
		formats []*jsonata.Expr
	}
)

var (
	style = table.Style{
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
			DrawBorder:      true,
			SeparateColumns: true,
			SeparateFooter:  false,
			SeparateHeader:  false,
			SeparateRows:    false,
		},
		Title: table.TitleOptionsDefault,
	}
)

func NewTable(columns []TableColumnDefinition) (*TablePrinter, error) {
	names := make([]string, len(columns))
	formats := make([]*jsonata.Expr, len(columns))

	for i, column := range columns {
		name := strcase.ToScreamingSnake(column.Name)
		format, err := jsonata.Compile(column.Format)
		if err != nil {
			return nil, err
		}

		names[i] = name
		formats[i] = format
	}

	return &TablePrinter{
		names:   names,
		formats: formats,
	}, nil
}

func (p *TablePrinter) Print(data any) (string, error) {
	value, err := primitive.MarshalText(data)
	if err != nil {
		return "", err
	}
	values, ok := value.(*primitive.Slice)
	if !ok {
		return "", nil
	}
	elements := values.Slice()

	header := make(table.Row, len(p.names))
	for i, name := range p.names {
		header[i] = name
	}

	rows := make([]table.Row, len(elements))
	for i, element := range elements {
		row := make(table.Row, len(p.formats))
		for j, format := range p.formats {
			if data, err := format.Eval(element); err != nil {
				row[j] = nil
			} else {
				row[j] = data
			}
		}
		rows[i] = row
	}

	tb := table.NewWriter()
	tb.SetStyle(style)
	tb.AppendHeader(header)
	tb.AppendRows(rows)

	return tb.Render(), nil
}
