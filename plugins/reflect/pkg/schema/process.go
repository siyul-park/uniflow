package schema

import (
	"context"
	"encoding/json"

	"github.com/siyul-park/sqlbridge/schema"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/xwb1989/sqlparser"
)

// ProcessTable represents a table of processes from a runtime agent.
type ProcessTable struct {
	agent *runtime.Agent
}

// NewProcessTable creates a new ProcessTable with the given agent.
func NewProcessTable(agent *runtime.Agent) *ProcessTable {
	return &ProcessTable{agent: agent}
}

// Scan returns a cursor for the processes in the agent, formatted as rows.
func (t *ProcessTable) Scan(_ context.Context) (schema.Cursor, error) {
	var rows []schema.Row
	for _, proc := range t.agent.Processes() {
		raw, err := json.Marshal(proc)
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
