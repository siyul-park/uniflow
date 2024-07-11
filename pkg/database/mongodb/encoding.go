package mongodb

import (
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	toLowerCamel = changeCase(strcase.ToLowerCamel)
	toSnake      = changeCase(strcase.ToSnake)
)

func filterToBson(filter *database.Filter) (bson.D, error) {
	if filter == nil {
		return bson.D{}, nil
	}

	switch filter.OP {
	case database.AND, database.OR:
		var values bson.A
		for _, child := range filter.Children {
			value, err := filterToBson(child)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}

		op := "$and"
		if filter.OP == database.OR {
			op = "$or"
		}
		return bson.D{{Key: op, Value: values}}, nil

	case database.NULL, database.NNULL:
		k := internalKey(filter.Key)
		op := "$eq"
		if filter.OP == database.NNULL {
			op = "$ne"
		}
		return bson.D{{Key: k, Value: bson.M{op: nil}}}, nil

	default:
		k := internalKey(filter.Key)
		v, err := toBson(filter.Value)
		if err != nil {
			return nil, err
		}

		ops := map[database.OP]string{
			database.EQ:  "$eq",
			database.NE:  "$ne",
			database.LT:  "$lt",
			database.LTE: "$lte",
			database.GT:  "$gt",
			database.GTE: "$gte",
			database.IN:  "$in",
			database.NIN: "$nin",
		}
		op, ok := ops[filter.OP]
		if !ok {
			return nil, errors.WithStack(encoding.ErrUnsupportedValue)
		}
		return bson.D{{Key: k, Value: bson.M{op: v}}}, nil
	}
}

func bsonToFilter(data interface{}, filter **database.Filter) error {
	raw, ok := bsonMA(data)
	if !ok {
		return errors.WithStack(encoding.ErrUnsupportedValue)
	}

	var children []*database.Filter
	for _, curr := range raw {
		for key, value := range curr {
			switch key {
			case "$and", "$or":
				if filters, ok := bsonMA(value); ok {
					var op database.OP
					if key == "$and" {
						op = database.AND
					} else if key == "$or" {
						op = database.OR
					}

					var childs []*database.Filter
					for _, filterData := range filters {
						var child *database.Filter
						if err := bsonToFilter(filterData, &child); err != nil {
							return err
						}
						childs = append(childs, child)
					}

					children = append(children, &database.Filter{
						OP:       op,
						Children: childs,
					})
				} else {
					return errors.WithStack(encoding.ErrUnsupportedValue)
				}

			case "$not":
				var child *database.Filter
				if err := bsonToFilter(value, &child); err != nil {
					return err
				}
				switch child.OP {
				case database.EQ:
					child.OP = database.NE
				case database.NE:
					child.OP = database.EQ
				case database.IN:
					child.OP = database.NIN
				case database.NIN:
					child.OP = database.IN
				case database.NULL:
					child.OP = database.NNULL
				case database.NNULL:
					child.OP = database.NULL
				}

				children = append(children, child)

			default:
				if value, ok := bsonM(value); ok {
					for k, v := range value {
						if !strings.HasPrefix(k, "$") {
							return errors.WithStack(encoding.ErrUnsupportedValue)
						}

						var op database.OP
						switch k {
						case "$eq":
							if v == nil {
								op = database.NULL
							} else {
								op = database.EQ
							}
						case "$ne":
							if v == nil {
								op = database.NNULL
							} else {
								op = database.NE
							}
						case "$lt":
							op = database.LT
						case "$lte":
							op = database.LTE
						case "$gt":
							op = database.GT
						case "$gte":
							op = database.GTE
						case "$in":
							op = database.IN
						case "$nin":
							op = database.NIN
						default:
							return errors.WithStack(encoding.ErrUnsupportedValue)
						}

						var value types.Value
						if err := fromBson(v, &value); err != nil {
							return err
						}

						children = append(children, &database.Filter{
							Key:   externalKey(key),
							OP:    op,
							Value: value,
						})
					}
				} else {
					return errors.WithStack(encoding.ErrUnsupportedValue)
				}
			}
		}
	}

	switch len(children) {
	case 0:
		*filter = nil
	case 1:
		*filter = children[0]
	default:
		*filter = &database.Filter{
			OP:       database.AND,
			Children: children,
		}
	}

	return nil
}

func toBson(data types.Value) (any, error) {
	if data == nil {
		return primitive.Null{}, nil
	}

	if s, ok := data.(types.Map); ok {
		t := make(primitive.M, s.Len())
		for _, k := range s.Keys() {
			v, _ := s.Get(k)
			if k, ok := k.(types.String); !ok {
				return nil, errors.WithStack(encoding.ErrUnsupportedValue)
			} else {
				if v, err := toBson(v); err != nil {
					return nil, err
				} else {
					t[internalKey(k.String())] = v
				}
			}
		}
		return t, nil
	} else if s, ok := data.(types.Slice); ok {
		t := make(primitive.A, s.Len())
		for i := 0; i < s.Len(); i++ {
			if v, err := toBson(s.Get(i)); err != nil {
				return nil, err
			} else {
				t[i] = v
			}
		}
		return t, nil
	} else {
		return data.Interface(), nil
	}
}

func fromBson(data any, v *types.Value) error {
	if data == nil {
		*v = nil
		return nil
	} else if _, ok := data.(primitive.Null); ok {
		*v = nil
		return nil
	} else if _, ok := data.(primitive.Undefined); ok {
		*v = nil
		return nil
	} else if s, ok := data.(primitive.Binary); ok {
		*v = types.NewBinary(s.Data)
		return nil
	} else if s, ok := data.(primitive.A); ok {
		values := make([]types.Value, len(s))
		for i, e := range s {
			var value types.Value
			if err := fromBson(e, &value); err != nil {
				return err
			}
			values[i] = value
		}
		*v = types.NewSlice(values...)
		return nil
	} else if s, ok := bsonM(data); ok {
		pairs := make([]types.Value, len(s)*2)
		i := 0
		for k, v := range s {
			var value types.Value
			if err := fromBson(v, &value); err != nil {
				return err
			}
			pairs[i*2] = types.NewString(externalKey(k))
			pairs[i*2+1] = value
			i += 1
		}
		*v = types.NewMap(pairs...)
		return nil
	} else if s, err := types.BinaryEncoder.Encode(data); err == nil {
		*v = s
		return nil
	}
	return errors.WithStack(encoding.ErrUnsupportedValue)
}

func sortToBson(sorts []database.Sort) bson.D {
	sort := bson.D{}
	for _, s := range sorts {
		sort = append(sort, bson.E{
			Key:   internalKey(s.Key),
			Value: orderToInt(s.Order),
		})
	}
	return sort
}

func orderToInt(order database.Order) int {
	if order == database.OrderASC {
		return 1
	}
	return -1
}

func internalKey(key string) string {
	if key == "id" {
		return "_id"
	}
	return toLowerCamel(key)
}

func externalKey(key string) string {
	if key == "_id" {
		return "id"
	}
	return toSnake(key)
}

func bsonMA(value any) ([]bson.M, bool) {
	if m, ok := bsonM(value); ok {
		return []bson.M{m}, true
	}

	var m []bson.M
	if v, ok := value.(primitive.A); ok {
		for _, e := range v {
			if e, ok := bsonM(e); ok {
				m = append(m, e)
			} else {
				return nil, false
			}
		}
	}

	return m, true
}

func bsonM(value any) (bson.M, bool) {
	if v, ok := value.(bson.M); ok {
		return v, true
	} else if v, ok := value.(bson.D); ok {
		m := make(bson.M, len(v))
		for _, e := range v {
			m[e.Key] = e.Value
		}
		return m, true
	} else if v, ok := value.(primitive.E); ok {
		return bson.M{v.Key: v.Value}, true
	}
	return nil, false
}

func changeCase(convert func(string) string) func(string) string {
	return func(str string) string {
		var tokens []string
		for _, curr := range strings.Split(str, ".") {
			tokens = append(tokens, convert(curr))
		}
		return strings.Join(tokens, ".")
	}
}
