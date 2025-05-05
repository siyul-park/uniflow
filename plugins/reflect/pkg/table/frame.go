package table

import (
	"context"
	"encoding/json"

	"github.com/siyul-park/sqlbridge/schema"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/xwb1989/sqlparser"
)

// FrameTable represents a table of frames from a runtime agent.
type FrameTable struct {
	agent *runtime.Agent
}

// NewFrameTable creates a new FrameTable with the given agent.
func NewFrameTable(agent *runtime.Agent) *FrameTable {
	return &FrameTable{agent: agent}
}

// Scan returns a cursor for the frames in the agent, formatted as rows.
func (t *FrameTable) Scan(_ context.Context) (schema.Cursor, error) {
	var rows []schema.Row
	for _, proc := range t.agent.Processes() {
		for _, frm := range t.agent.Frames(proc.ID()) {
			raw, err := json.Marshal(frm)
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
	}
	return schema.NewInMemoryCursor(rows), nil
}
