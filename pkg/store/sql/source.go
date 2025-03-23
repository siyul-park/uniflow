package sql

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/araddon/qlbridge/schema"
	"github.com/araddon/qlbridge/value"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/siyul-park/uniflow/pkg/types"
)

type SourceOptions struct {
	Timeout time.Duration
	Limit   int
}

type Source struct {
	source  store.Source
	schema  *schema.Schema
	tables  map[string]*schema.Table
	timeout time.Duration
	limit   int
	mu      sync.Mutex
}

var _ schema.Source = (*Source)(nil)

func NewSource(source store.Source, opts ...SourceOptions) *Source {
	var timeout time.Duration
	var limit = 16
	for _, opt := range opts {
		if opt.Timeout > 0 {
			timeout = opt.Timeout
		}
		if opt.Limit > 0 {
			limit = opt.Limit
		}
	}

	return &Source{
		source:  source,
		tables:  make(map[string]*schema.Table),
		timeout: timeout,
		limit:   limit,
	}
}

func (s *Source) Init() {
}

func (s *Source) Setup(ss *schema.Schema) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.schema = ss
	return nil
}

func (s *Source) Open(source string) (schema.Conn, error) {
	st, err := s.source.Open(source)
	if err != nil {
		return nil, err
	}
	tbl, err := s.Table(source)
	if err != nil {
		return nil, err
	}
	return &connection{store: st, table: tbl}, nil
}

func (s *Source) Tables() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	tables := make([]string, 0, len(s.tables))
	for table := range s.tables {
		tables = append(tables, table)
	}
	return tables
}

func (s *Source) Table(table string) (*schema.Table, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	tbl, ok := s.tables[table]
	if !ok {
		st, err := s.source.Open(table)
		if err != nil {
			return nil, err
		}

		cursor, err := st.Find(ctx, nil, store.FindOptions{Limit: s.limit})
		if err != nil {
			return nil, err
		}

		var rows []types.Map
		if err := cursor.All(ctx, &rows); err != nil {
			return nil, err
		}

		tbl = schema.NewTable(table)

		var cols []string
		for _, row := range rows {
			for key, val := range row.Range() {
				field, err := types.Cast[string](key)
				if err != nil {
					return nil, err
				}
				field = strings.ToLower(field)

				if tbl.HasField(field) {
					continue
				}

				switch val.(type) {
				case types.Binary:
					tbl.AddField(schema.NewFieldBase(field, value.ByteSliceType, -1, "[]byte"))
				case types.Boolean:
					tbl.AddField(schema.NewFieldBase(field, value.BoolType, 1, "bool"))
				case types.Float32:
					tbl.AddField(schema.NewFieldBase(field, value.NumberType, 32, "float32"))
				case types.Float64:
					tbl.AddField(schema.NewFieldBase(field, value.NumberType, 64, "float64"))
				case types.Int:
					tbl.AddField(schema.NewFieldBase(field, value.NumberType, 32, "int"))
				case types.Int8:
					tbl.AddField(schema.NewFieldBase(field, value.NumberType, 8, "int8"))
				case types.Int16:
					tbl.AddField(schema.NewFieldBase(field, value.NumberType, 16, "int16"))
				case types.Int32:
					tbl.AddField(schema.NewFieldBase(field, value.NumberType, 32, "int32"))
				case types.Int64:
					tbl.AddField(schema.NewFieldBase(field, value.NumberType, 64, "int64"))
				case types.Uint:
					tbl.AddField(schema.NewFieldBase(field, value.NumberType, 32, "uint"))
				case types.Uint8:
					tbl.AddField(schema.NewFieldBase(field, value.NumberType, 8, "uint8"))
				case types.Uint16:
					tbl.AddField(schema.NewFieldBase(field, value.NumberType, 16, "uint16"))
				case types.Uint32:
					tbl.AddField(schema.NewFieldBase(field, value.NumberType, 32, "uint32"))
				case types.Uint64:
					tbl.AddField(schema.NewFieldBase(field, value.NumberType, 64, "uint64"))
				case types.Map:
					tbl.AddField(schema.NewFieldBase(field, value.MapValueType, 24, "map[string]interface{}"))
				case types.Slice:
					tbl.AddField(schema.NewFieldBase(field, value.SliceValueType, -1, "[]interface{}"))
				case types.String:
					tbl.AddField(schema.NewFieldBase(field, value.StringType, 255, "string"))
				default:
					continue
				}

				cols = append(cols, field)
			}
		}

		tbl.SetColumns(cols)

		s.tables[table] = tbl
	}
	return tbl, nil
}

func (s *Source) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.schema = nil
	s.tables = make(map[string]*schema.Table)
	return nil
}
