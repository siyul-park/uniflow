package database

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/siyul-park/uniflow/pkg/primitive"
)

// Filter is a filter for finding matched primitive.
type Filter struct {
	OP       OP
	Key      string
	Value    primitive.Value
	Children []*Filter
}

// OP represents comparison operators for filters.
type OP string

const (
	EQ    OP = "="
	NE    OP = "!="
	LT    OP = "<"
	LTE   OP = "<="
	GT    OP = ">"
	GTE   OP = ">="
	IN    OP = "IN"
	NIN   OP = "NOT IN"
	NULL  OP = "IS NULL"
	NNULL OP = "IS NOT NULL"
	AND   OP = "AND"
	OR    OP = "OR"
)

// filterHelper is a helper for building filters.
type filterHelper struct {
	key string
}

// Where creates a filterHelper for a specific key.
func Where(key string) *filterHelper {
	return &filterHelper{
		key: key,
	}
}

// EQ creates an equality filter.
func (fh *filterHelper) EQ(value primitive.Value) *Filter {
	return &Filter{
		OP:    EQ,
		Key:   fh.key,
		Value: value,
	}
}

// NE creates a not-equal filter.
func (fh *filterHelper) NE(value primitive.Value) *Filter {
	return &Filter{
		OP:    NE,
		Key:   fh.key,
		Value: value,
	}
}

// LT creates a less-than filter.
func (fh *filterHelper) LT(value primitive.Value) *Filter {
	return &Filter{
		Key:   fh.key,
		OP:    LT,
		Value: value,
	}
}

// LTE creates a less-than-or-equal filter.
func (fh *filterHelper) LTE(value primitive.Value) *Filter {
	return &Filter{
		OP:    LTE,
		Key:   fh.key,
		Value: value,
	}
}

// GT creates a greater-than filter.
func (fh *filterHelper) GT(value primitive.Value) *Filter {
	return &Filter{
		OP:    GT,
		Key:   fh.key,
		Value: value,
	}
}

// GTE creates a greater-than-or-equal filter.
func (fh *filterHelper) GTE(value primitive.Value) *Filter {
	return &Filter{
		OP:    GTE,
		Key:   fh.key,
		Value: value,
	}
}

// IN creates an "in" filter.
func (fh *filterHelper) IN(slice ...primitive.Value) *Filter {
	return &Filter{
		OP:    IN,
		Key:   fh.key,
		Value: primitive.NewSlice(slice...),
	}
}

// NotIN creates a "not in" filter.
func (fh *filterHelper) NotIN(slice ...primitive.Value) *Filter {
	return &Filter{
		OP:    NIN,
		Key:   fh.key,
		Value: primitive.NewSlice(slice...),
	}
}

// IsNull creates an "is null" filter.
func (fh *filterHelper) IsNull() *Filter {
	return &Filter{
		OP:  NULL,
		Key: fh.key,
	}
}

// IsNotNull creates an "is not null" filter.
func (fh *filterHelper) IsNotNull() *Filter {
	return &Filter{
		OP:  NNULL,
		Key: fh.key,
	}
}

// And creates an "and" filter.
func (ft *Filter) And(x ...*Filter) *Filter {
	var v []*Filter
	for _, e := range append([]*Filter{ft}, x...) {
		if e != nil {
			v = append(v, e)
		}
	}

	return &Filter{
		OP:       AND,
		Children: v,
	}
}

// Or creates an "or" filter.
func (ft *Filter) Or(x ...*Filter) *Filter {
	var v []*Filter
	for _, e := range append([]*Filter{ft}, x...) {
		if e != nil {
			v = append(v, e)
		}
	}

	return &Filter{
		OP:       OR,
		Children: v,
	}
}

// String converts a filter to a string representation.
func (ft *Filter) String() (string, error) {
	if ft.OP == AND || ft.OP == OR {
		var parsed []string
		for _, child := range ft.Children {
			if c, err := child.String(); err != nil {
				return "", err
			} else {
				parsed = append(parsed, "("+c+")")
			}
		}
		return strings.Join(parsed, " "+string(ft.OP)+" "), nil
	}
	if ft.OP == NULL || ft.OP == NNULL {
		return ft.Key + " " + string(ft.OP), nil
	}

	b, err := json.Marshal(primitive.Interface(ft.Value))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %s %s", ft.Key, string(ft.OP), string(b)), nil
}
