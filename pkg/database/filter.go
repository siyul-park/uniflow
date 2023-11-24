package database

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/siyul-park/uniflow/pkg/primitive"
)

type (
	// Filter is a filter for find matched primitive.
	Filter struct {
		OP    OP
		Key   string
		Value any
	}

	filterHelper struct {
		key string
	}

	OP string
)

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

func Where(key string) *filterHelper {
	return &filterHelper{
		key: key,
	}
}

func (fh *filterHelper) EQ(value primitive.Object) *Filter {
	return &Filter{
		OP:    EQ,
		Key:   fh.key,
		Value: value,
	}
}

func (fh *filterHelper) NE(value primitive.Object) *Filter {
	return &Filter{
		OP:    NE,
		Key:   fh.key,
		Value: value,
	}
}

func (fh *filterHelper) LT(value primitive.Object) *Filter {
	return &Filter{
		Key:   fh.key,
		OP:    LT,
		Value: value,
	}
}

func (fh *filterHelper) LTE(value primitive.Object) *Filter {
	return &Filter{
		OP:    LTE,
		Key:   fh.key,
		Value: value,
	}
}

func (fh *filterHelper) GT(value primitive.Object) *Filter {
	return &Filter{
		OP:    GT,
		Key:   fh.key,
		Value: value,
	}
}

func (fh *filterHelper) GTE(value primitive.Object) *Filter {
	return &Filter{
		OP:    GTE,
		Key:   fh.key,
		Value: value,
	}
}

func (fh *filterHelper) IN(slice ...primitive.Object) *Filter {
	return &Filter{
		OP:    IN,
		Key:   fh.key,
		Value: primitive.NewSlice(slice...),
	}
}

func (fh *filterHelper) NotIN(slice ...primitive.Object) *Filter {
	return &Filter{
		OP:    NIN,
		Key:   fh.key,
		Value: primitive.NewSlice(slice...),
	}
}

func (fh *filterHelper) IsNull() *Filter {
	return &Filter{
		OP:  NULL,
		Key: fh.key,
	}
}

func (fh *filterHelper) IsNotNull() *Filter {
	return &Filter{
		OP:  NNULL,
		Key: fh.key,
	}
}

func (ft *Filter) And(x ...*Filter) *Filter {
	var v []*Filter
	for _, e := range append([]*Filter{ft}, x...) {
		if e != nil {
			v = append(v, e)
		}
	}

	return &Filter{
		OP:    AND,
		Value: v,
	}
}

func (ft *Filter) Or(x ...*Filter) *Filter {
	var v []*Filter
	for _, e := range append([]*Filter{ft}, x...) {
		if e != nil {
			v = append(v, e)
		}
	}

	return &Filter{
		OP:    OR,
		Value: v,
	}
}

func (ft *Filter) String() (string, error) {
	if ft.OP == AND || ft.OP == OR {
		var parsed []string
		if value, ok := ft.Value.([]*Filter); ok {
			for _, v := range value {
				c, e := v.String()
				if e != nil {
					return "", e
				}
				parsed = append(parsed, "("+c+")")
			}
		}
		return strings.Join(parsed, " "+string(ft.OP)+" "), nil
	}
	if ft.OP == NULL || ft.OP == NNULL {
		return ft.Key + " " + string(ft.OP), nil
	}

	v, _ := ft.Value.(primitive.Object)
	b, err := json.Marshal(primitive.Interface(v))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %s %s", ft.Key, string(ft.OP), string(b)), nil
}
