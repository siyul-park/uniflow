package sql

import (
	"context"
	sqldriver "database/sql/driver"
	"fmt"
	"github.com/araddon/qlbridge/datasource"
	"github.com/araddon/qlbridge/exec"
	"github.com/araddon/qlbridge/expr"
	"github.com/araddon/qlbridge/plan"
	"github.com/araddon/qlbridge/schema"
	"github.com/pkg/errors"
	"io"
	"strconv"
	"strings"
	"time"
)

type Connector struct {
	applyer  schema.Applyer
	registry *schema.Registry
}

type driver struct {
	registry *schema.Registry
}

type conn struct {
	schema *schema.Schema
}

type stmt struct {
	schema *schema.Schema
	query  string
}

type rows struct {
	*exec.TaskBase
	columns []string
}

var _ sqldriver.Connector = (*Connector)(nil)
var _ sqldriver.Driver = (*driver)(nil)
var _ sqldriver.Conn = (*conn)(nil)
var _ sqldriver.ExecerContext = (*conn)(nil)
var _ sqldriver.QueryerContext = (*conn)(nil)
var _ sqldriver.StmtExecContext = (*stmt)(nil)
var _ sqldriver.StmtQueryContext = (*stmt)(nil)
var _ sqldriver.Rows = (*rows)(nil)

func NewConnector() *Connector {
	applyer := schema.NewApplyer(datasource.SchemaDBStoreProvider)
	registry := schema.NewRegistry(applyer)

	applyer.Init(registry)

	return &Connector{
		applyer:  applyer,
		registry: registry,
	}
}

func (c *Connector) RegisterSourceAsSchema(name string, source schema.Source) error {
	s := schema.NewSchemaSource(name, source)

	source.Init()
	if err := source.Setup(s); err != nil {
		return err
	}

	if err := c.registry.SchemaAdd(s); err != nil {
		return err
	}

	if err := s.DS.Setup(s); err != nil {
		return err
	}

	for _, n := range s.Tables() {
		tbl, err := s.Table(n)
		if err != nil || tbl == nil {
			continue
		}
		if err := c.applyer.AddOrUpdateOnSchema(s, tbl); err != nil {
			return err
		}
	}
	return nil
}

func (c *Connector) Connect(_ context.Context) (sqldriver.Conn, error) {
	for _, name := range c.registry.Schemas() {
		s, ok := c.registry.Schema(name)
		if !ok || s == nil {
			return nil, fmt.Errorf("no schema was found for %q", name)
		}
		return &conn{schema: s}, nil
	}
	return nil, fmt.Errorf("no schema was found for %q", c.registry)
}

func (c *Connector) Driver() sqldriver.Driver {
	return &driver{registry: c.registry}
}

func (d *driver) Open(name string) (sqldriver.Conn, error) {
	s, ok := d.registry.Schema(name)
	if !ok || s == nil {
		return nil, fmt.Errorf("no schema was found for %q", name)
	}
	return &conn{schema: s}, nil
}

func (c *conn) Prepare(_ string) (sqldriver.Stmt, error) {
	return nil, expr.ErrNotImplemented
}

func (c *conn) Begin() (sqldriver.Tx, error) {
	return nil, expr.ErrNotImplemented
}

func (c *conn) QueryContext(ctx context.Context, query string, args []sqldriver.NamedValue) (sqldriver.Rows, error) {
	s := &stmt{schema: c.schema, query: query}
	return s.QueryContext(ctx, args)
}

func (c *conn) ExecContext(ctx context.Context, query string, args []sqldriver.NamedValue) (sqldriver.Result, error) {
	s := &stmt{schema: c.schema, query: query}
	return s.ExecContext(ctx, args)
}

func (c *conn) Close() error {
	return nil
}

func (s *stmt) NumInput() int {
	return 0
}

func (s *stmt) QueryContext(ctx context.Context, args []sqldriver.NamedValue) (sqldriver.Rows, error) {
	query, err := s.format(s.query, args)
	if err != nil {
		return nil, err
	}

	pctx := plan.NewContext(query)
	pctx.Schema = s.schema
	pctx.Context = ctx

	job, err := exec.BuildSqlJob(pctx)
	if err != nil {
		return nil, err
	}

	stepper := exec.NewTaskStepper(pctx)

	columns := pctx.Projection.Stmt.Columns.AliasedFieldNames()
	if pctx.Projection.Proj != nil {
		columns = nil
		for _, col := range pctx.Projection.Proj.Columns {
			columns = append(columns, col.As)
		}
	}

	r := &rows{TaskBase: stepper.TaskBase, columns: columns}
	r.Handler = func(ctx *plan.Context, msg schema.Message) bool {
		select {
		case r.MessageOut() <- msg:
			return true
		case <-r.SigChan():
			return false
		}
	}

	if err := job.RootTask.Add(r); err != nil {
		return nil, err
	}

	if err := job.Setup(); err != nil {
		return nil, err
	}

	go func() {
		_ = job.Run()
		_ = job.Close()
	}()

	return r, nil
}

func (s *stmt) ExecContext(ctx context.Context, args []sqldriver.NamedValue) (sqldriver.Result, error) {
	//TODO implement me
	panic("implement me")
}

func (s *stmt) Close() error {
	return nil
}

func (s *stmt) format(query string, args []sqldriver.NamedValue) (string, error) {
	if len(args) == 0 || strings.ContainsAny(query, `'"`) {
		return query, nil
	}

	var result string
	for _, arg := range args {
		var placeholder string
		if arg.Name != "" {
			placeholder = fmt.Sprintf(":%s", arg.Name)
		} else {
			placeholder = "?"
		}

		i := strings.Index(query, placeholder)
		if i == -1 {
			return "", errors.New("number of parameters doesn't match number of placeholders")
		}

		var str string
		switch v := arg.Value.(type) {
		case nil:
			str = "NULL"
		case string:
			str = "'" + strings.ReplaceAll(v, "'", "''") + "'"
		case []byte:
			str = "'" + strings.ReplaceAll(string(v), "'", "''") + "'"
		case int64:
			str = strconv.FormatInt(v, 10)
		case time.Time:
			str = "'" + v.Format("2006-01-02 15:04:05.000000000") + "'"
		case bool:
			if v {
				str = "1"
			} else {
				str = "0"
			}
		case float64:
			str = strconv.FormatFloat(v, 'e', 12, 64)
		default:
			return "", fmt.Errorf("%v (%T) can't be handled", v, v)
		}

		result += query[:i] + str
		query = query[i+len(placeholder):]
	}

	result += query
	return result, nil
}

func (r *rows) Columns() []string {
	return r.columns
}

func (r *rows) Next(dest []sqldriver.Value) error {
	select {
	case <-r.SigChan():
		return exec.ErrShuttingDown
	case err := <-r.ErrChan():
		return err
	case msg, ok := <-r.MessageIn():
		if !ok || msg == nil {
			return io.EOF
		}
		if mt, ok := msg.Body().(*datasource.SqlDriverMessageMap); ok {
			for i := 0; i < len(dest); i++ {
				if i < len(mt.Values()) {
					dest[i] = mt.Values()[i]
				}
			}
		}
	}
	return nil
}

func (r *rows) Close() error {
	return r.TaskBase.Close()
}
