package table

import (
	"context"
	"encoding/json"

	"github.com/siyul-park/sqlbridge/schema"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/xwb1989/sqlparser"
)

// SymbolTable represents a table of symbols from a runtime agent.
type SymbolTable struct {
	agent *runtime.Agent
}

// NewSymbolTable creates a new SymbolTable with the given agent.
func NewSymbolTable(agent *runtime.Agent) *SymbolTable {
	return &SymbolTable{agent: agent}
}

// Scan returns a cursor for the symbols in the agent, formatted as rows.
func (t *SymbolTable) Scan(_ context.Context) (schema.Cursor, error) {
	var rows []schema.Row
	for _, sb := range t.agent.Symbols() {
		raw, err := json.Marshal(sb)
		if err != nil {
			return nil, err
		}

		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return nil, err
		}

		var row schema.Row
		for k, v := range data {
			col := &sqlparser.ColName{Name: sqlparser.NewColIdent(k)}
			val, err := schema.Marshal(v)
			if err != nil {
				return nil, err
			}
			row.Columns = append(row.Columns, col)
			row.Values = append(row.Values, val)
		}
		rows = append(rows, row)
	}
	return schema.NewInMemoryCursor(rows), nil
}
