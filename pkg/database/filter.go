package database

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/siyul-park/uniflow/pkg/object"
)

// Filter represents a filter used to find matching primitives.
type Filter struct {
	OP       OP            // Comparison operator for the filter.
	Key      string        // Key to apply the filter on.
	Value    object.Object // Value to compare against.
	Children []*Filter     // Child filters for logical operations.
}

// OP represents comparison operators for filters.
type OP string

const (
	EQ    OP = "="           // Equal comparison operator.
	NE    OP = "!="          // Not equal comparison operator.
	LT    OP = "<"           // Less than comparison operator.
	LTE   OP = "<="          // Less than or equal comparison operator.
	GT    OP = ">"           // Greater than comparison operator.
	GTE   OP = ">="          // Greater than or equal comparison operator.
	IN    OP = "IN"          // In comparison operator.
	NIN   OP = "NOT IN"      // Not in comparison operator.
	NULL  OP = "IS NULL"     // Is null comparison operator.
	NNULL OP = "IS NOT NULL" // Is not null comparison operator.
	AND   OP = "AND"         // Logical AND operator.
	OR    OP = "OR"          // Logical OR operator.
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

// Equal creates an equality filter.
func (fh *filterHelper) Equal(value object.Object) *Filter {
	return &Filter{
		OP:    EQ,
		Key:   fh.key,
		Value: value,
	}
}

// NotEqual creates a not-equal filter.
func (fh *filterHelper) NotEqual(value object.Object) *Filter {
	return &Filter{
		OP:    NE,
		Key:   fh.key,
		Value: value,
	}
}

// LessThan creates a less-than filter.
func (fh *filterHelper) LessThan(value object.Object) *Filter {
	return &Filter{
		OP:    LT,
		Key:   fh.key,
		Value: value,
	}
}

// LessThanOrEqual creates a less-than-or-equal filter.
func (fh *filterHelper) LessThanOrEqual(value object.Object) *Filter {
	return &Filter{
		OP:    LTE,
		Key:   fh.key,
		Value: value,
	}
}

// GreaterThan creates a greater-than filter.
func (fh *filterHelper) GreaterThan(value object.Object) *Filter {
	return &Filter{
		OP:    GT,
		Key:   fh.key,
		Value: value,
	}
}

// GreaterThanOrEqual creates a greater-than-or-equal filter.
func (fh *filterHelper) GreaterThanOrEqual(value object.Object) *Filter {
	return &Filter{
		OP:    GTE,
		Key:   fh.key,
		Value: value,
	}
}

// In creates an in filter.
func (fh *filterHelper) In(slice ...object.Object) *Filter {
	return &Filter{
		OP:    IN,
		Key:   fh.key,
		Value: object.NewSlice(slice...),
	}
}

// NotIn creates a not-in filter.
func (fh *filterHelper) NotIn(slice ...object.Object) *Filter {
	return &Filter{
		OP:    NIN,
		Key:   fh.key,
		Value: object.NewSlice(slice...),
	}
}

// IsNull creates an is-null filter.
func (fh *filterHelper) IsNull() *Filter {
	return &Filter{
		OP:  NULL,
		Key: fh.key,
	}
}

// IsNotNull creates an is-no-null filter.
func (fh *filterHelper) IsNotNull() *Filter {
	return &Filter{
		OP:  NNULL,
		Key: fh.key,
	}
}

// And creates an and filter.
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

// Or creates an or filter.
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

// String converts a filter to its string representation.
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

	b, err := json.Marshal(object.Interface(ft.Value))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %s %s", ft.Key, string(ft.OP), string(b)), nil
}
