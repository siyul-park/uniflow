package mongodb

import (
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"go.mongodb.org/mongo-driver/bson"
	bsonprimitive "go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	toLowerCamel = changeCase(strcase.ToLowerCamel)
	toSnake      = changeCase(strcase.ToSnake)
)

func changeCase(convert func(string) string) func(string) string {
	return func(str string) string {
		var tokens []string
		for _, curr := range strings.Split(str, ".") {
			tokens = append(tokens, convert(curr))
		}
		return strings.Join(tokens, ".")
	}
}

func marshalFilter(filter *database.Filter) (any, error) {
	if filter == nil {
		return bson.D{}, nil
	}

	switch filter.OP {
	case database.AND, database.OR:
		var values bson.A
		for _, e := range filter.Children {
			if value, err := marshalFilter(e); err != nil {
				return nil, err
			} else {
				values = append(values, value)
			}
		}

		if filter.OP == database.AND {
			return bson.D{{Key: "$and", Value: values}}, nil
		} else if filter.OP == database.OR {
			return bson.D{{Key: "$or", Value: values}}, nil
		}
	case database.NULL, database.NNULL:
		k := marshalKey(filter.Key)

		if filter.OP == database.NULL {
			return bson.D{{Key: k, Value: bson.M{"$eq": nil}}}, nil
		} else if filter.OP == database.NNULL {
			return bson.D{{Key: k, Value: bson.M{"$ne": nil}}}, nil
		}
	default:
		k := marshalKey(filter.Key)
		v, err := marshalDocument(filter.Value)
		if err != nil {
			return nil, err
		}

		if filter.OP == database.EQ {
			return bson.D{{Key: k, Value: bson.M{"$eq": v}}}, nil
		} else if filter.OP == database.NE {
			return bson.D{{Key: k, Value: bson.M{"$ne": v}}}, nil
		} else if filter.OP == database.LT {
			return bson.D{{Key: k, Value: bson.M{"$lt": v}}}, nil
		} else if filter.OP == database.LTE {
			return bson.D{{Key: k, Value: bson.M{"$lte": v}}}, nil
		} else if filter.OP == database.GT {
			return bson.D{{Key: k, Value: bson.M{"$gt": v}}}, nil
		} else if filter.OP == database.GTE {
			return bson.D{{Key: k, Value: bson.M{"$gte": v}}}, nil
		} else if filter.OP == database.IN {
			return bson.D{{Key: k, Value: bson.M{"$in": v}}}, nil
		} else if filter.OP == database.NIN {
			return bson.D{{Key: k, Value: bson.M{"$nin": v}}}, nil
		}
	}

	return nil, errors.WithStack(encoding.ErrUnsupportedValue)
}

func unmarshalFilter(data any, filter **database.Filter) error {
	raw, ok := bsonMA(data)
	if !ok {
		return errors.WithStack(encoding.ErrInvalidValue)
	}

	var children []*database.Filter
	for _, curr := range raw {
		for key, value := range curr {
			if key == "$and" || key == "$or" {
				if value, ok := bsonMA(value); !ok {
					return errors.WithStack(encoding.ErrInvalidValue)
				} else {
					var values []*database.Filter
					for _, v := range value {
						var value *database.Filter
						if err := unmarshalFilter(v, &value); err != nil {
							return err
						}
						values = append(values, value)
					}

					if key == "$and" {
						children = append(children, &database.Filter{
							OP:       database.AND,
							Children: values,
						})
					} else if key == "$or" {
						children = append(children, &database.Filter{
							OP:       database.OR,
							Children: values,
						})
					}
				}
			} else if key == "$not" {
				var child *database.Filter
				if err := unmarshalFilter(value, &child); err != nil {
					return err
				}
				if child.OP == database.EQ {
					child.OP = database.NE
				} else if child.OP == database.NE {
					child.OP = database.EQ
				} else if child.OP == database.IN {
					child.OP = database.NIN
				} else if child.OP == database.NIN {
					child.OP = database.IN
				} else if child.OP == database.NULL {
					child.OP = database.NNULL
				} else if child.OP == database.NNULL {
					child.OP = database.NULL
				} else {
					return errors.WithStack(encoding.ErrInvalidValue)
				}
				children = append(children, child)
			} else if value, ok := bsonM(value); ok {
				for op, v := range value {
					if !strings.HasPrefix(op, "$") {
						return errors.WithStack(encoding.ErrInvalidValue)
					}
					child := &database.Filter{
						Key: unmarshalKey(key),
					}
					if op == "$eq" {
						if v == nil {
							child.OP = database.NULL
						} else {
							child.OP = database.EQ
						}
					} else if op == "$ne" {
						if v == nil {
							child.OP = database.NNULL
						} else {
							child.OP = database.NE
						}
					} else if op == "$lt" {
						child.OP = database.LT
					} else if op == "$lte" {
						child.OP = database.LTE
					} else if op == "$gt" {
						child.OP = database.GT
					} else if op == "$gte" {
						child.OP = database.GTE
					} else if op == "$in" {
						child.OP = database.IN
					} else if op == "$nin" {
						child.OP = database.NIN
					} else {
						return errors.WithStack(encoding.ErrInvalidValue)
					}

					var value primitive.Value
					if err := unmarshalDocument(v, &value); err != nil {
						return err
					}
					child.Value = value
					children = append(children, child)
				}
			} else {
				return errors.WithStack(encoding.ErrInvalidValue)
			}
		}
	}

	if len(children) == 0 {
		*filter = nil
	} else if len(children) == 1 {
		*filter = children[0]
	} else {
		*filter = &database.Filter{
			OP:       database.AND,
			Children: children,
		}
	}

	return nil
}

func marshalDocument(data primitive.Value) (any, error) {
	if data == nil {
		return bsonprimitive.Null{}, nil
	}

	if s, ok := data.(primitive.Binary); ok {
		return bsonprimitive.Binary{Data: []byte(s)}, nil
	} else if s, ok := data.(*primitive.Map); ok {
		t := make(bsonprimitive.M, s.Len())
		for _, k := range s.Keys() {
			v, _ := s.Get(k)
			if k, ok := k.(primitive.String); !ok {
				return nil, errors.WithStack(encoding.ErrInvalidValue)
			} else {
				if v, err := marshalDocument(v); err != nil {
					return nil, err
				} else {
					t[marshalKey(k.String())] = v
				}
			}
		}
		return t, nil
	} else if s, ok := data.(*primitive.Slice); ok {
		t := make(bsonprimitive.A, s.Len())
		for i := 0; i < s.Len(); i++ {
			if v, err := marshalDocument(s.Get(i)); err != nil {
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

func unmarshalDocument(data any, v *primitive.Value) error {
	if data == nil {
		*v = nil
		return nil
	} else if _, ok := data.(bsonprimitive.Null); ok {
		*v = nil
		return nil
	} else if _, ok := data.(bsonprimitive.Undefined); ok {
		*v = nil
		return nil
	} else if s, ok := data.(bsonprimitive.Binary); ok {
		*v = primitive.NewBinary(s.Data)
		return nil
	} else if s, ok := data.(bsonprimitive.A); ok {
		values := make([]primitive.Value, len(s))
		for i, e := range s {
			var value primitive.Value
			if err := unmarshalDocument(e, &value); err != nil {
				return err
			}
			values[i] = value
		}
		*v = primitive.NewSlice(values...)
		return nil
	} else if s, ok := bsonM(data); ok {
		pairs := make([]primitive.Value, len(s)*2)
		i := 0
		for k, v := range s {
			var value primitive.Value
			if err := unmarshalDocument(v, &value); err != nil {
				return err
			}
			pairs[i*2] = primitive.NewString(unmarshalKey(k))
			pairs[i*2+1] = value
			i += 1
		}
		*v = primitive.NewMap(pairs...)
		return nil
	} else if s, err := primitive.MarshalBinary(data); err == nil {
		*v = s
		return nil
	}
	return errors.WithStack(encoding.ErrInvalidValue)
}

func marshalSorts(sorts []database.Sort) bson.D {
	sort := bson.D{}
	for _, s := range sorts {
		sort = append(sort, bson.E{
			Key:   marshalKey(s.Key),
			Value: marshalOrder(s.Order),
		})
	}
	return sort
}

func marshalOrder(order database.Order) int {
	if order == database.OrderASC {
		return 1
	}
	return -1
}

func marshalKey(key string) string {
	if key == "id" {
		return "_id"
	}
	return toLowerCamel(key)
}

func unmarshalKey(key string) string {
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
	if v, ok := value.(bsonprimitive.A); ok {
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
	} else if v, ok := value.(bsonprimitive.E); ok {
		return bson.M{v.Key: v.Value}, true
	}
	return nil, false
}
