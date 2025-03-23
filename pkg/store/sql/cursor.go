package sql

import (
	sqldriver "database/sql/driver"
	"encoding/json"
	"github.com/araddon/qlbridge/datasource"
	"github.com/araddon/qlbridge/exec"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/siyul-park/uniflow/pkg/types"
	"strings"
)

type cursor struct {
	*exec.TaskBase
	cursor store.Cursor
}

var _ exec.TaskRunner = (*cursor)(nil)

func (c *cursor) Run() error {
	defer close(c.MessageOut())

	id := 0
	for c.cursor.Next(c.TaskBase.Ctx) {
		var row types.Map
		if err := c.cursor.Decode(&row); err != nil {
			return err
		}

		keys := make(map[string]int, row.Len())
		vals := make([]sqldriver.Value, 0, row.Len())

		for key, val := range row.Range() {
			field, err := types.Cast[string](key)
			if err != nil {
				return err
			}
			field = strings.ToLower(field)

			keys[field] = len(vals)

			switch v := val.(type) {
			case types.Map:
				val, err := json.Marshal(types.InterfaceOf(v))
				if err != nil {
					return err
				}
				vals = append(vals, val)
			default:
				vals = append(vals, types.InterfaceOf(v))
			}
		}

		id += 1
		msg := datasource.NewSqlDriverMessageMap(uint64(id), vals, keys)

		select {
		case <-c.SigChan():
			return nil
		case c.MessageOut() <- msg:
		}
	}
	return c.cursor.Close(c.TaskBase.Ctx)
}

func (c *cursor) Close() error {
	return c.cursor.Close(c.TaskBase.Ctx)
}
