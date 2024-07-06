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

func filterToBson(filter *database.Filter) (any, error) {
	if filter == nil {
		return bson.D{}, nil
	}

	switch filter.OP {
	case database.AND, database.OR:
		var values bson.A
		for _, e := range filter.Children {
			if value, err := filterToBson(e); err != nil {
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
		k := internalKey(filter.Key)

		if filter.OP == database.NULL {
			return bson.D{{Key: k, Value: bson.M{"$eq": nil}}}, nil
		} else if filter.OP == database.NNULL {
			return bson.D{{Key: k, Value: bson.M{"$ne": nil}}}, nil
		}
	default:
		k := internalKey(filter.Key)
		v, err := primitiveToBson(filter.Value)
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

	return nil, errors.WithStack(encoding.ErrInvalidArgument)
}

func bsonToFilter(data any, filter **database.Filter) error {
	raw, ok := bsonMA(data)
	if !ok {
		return errors.WithStack(encoding.ErrInvalidArgument)
	}

	var children []*database.Filter
	for _, curr := range raw {
		for key, value := range curr {
			if key == "$and" || key == "$or" {
				if value, ok := bsonMA(value); !ok {
					return errors.WithStack(encoding.ErrInvalidArgument)
				} else {
					var values []*database.Filter
					for _, v := range value {
						var value *database.Filter
						if err := bsonToFilter(v, &value); err != nil {
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
				if err := bsonToFilter(value, &child); err != nil {
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
					return errors.WithStack(encoding.ErrInvalidArgument)
				}
				children = append(children, child)
			} else if value, ok := bsonM(value); ok {
				for op, v := range value {
					if !strings.HasPrefix(op, "$") {
						return errors.WithStack(encoding.ErrInvalidArgument)
					}
					child := &database.Filter{
						Key: externalKey(key),
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
						return errors.WithStack(encoding.ErrInvalidArgument)
					}

					var value types.Object
					if err := bsonToPrimitive(v, &value); err != nil {
						return err
					}
					child.Value = value
					children = append(children, child)
				}
			} else {
				return errors.WithStack(encoding.ErrInvalidArgument)
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

func primitiveToBson(data types.Object) (any, error) {
	if data == nil {
		return primitive.Null{}, nil
	}

	if s, ok := data.(types.Map); ok {
		t := make(primitive.M, s.Len())
		for _, k := range s.Keys() {
			v, _ := s.Get(k)
			if k, ok := k.(types.String); !ok {
				return nil, errors.WithStack(encoding.ErrInvalidArgument)
			} else {
				if v, err := primitiveToBson(v); err != nil {
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
			if v, err := primitiveToBson(s.Get(i)); err != nil {
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

func bsonToPrimitive(data any, v *types.Object) error {
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
		values := make([]types.Object, len(s))
		for i, e := range s {
			var value types.Object
			if err := bsonToPrimitive(e, &value); err != nil {
				return err
			}
			values[i] = value
		}
		*v = types.NewSlice(values...)
		return nil
	} else if s, ok := bsonM(data); ok {
		pairs := make([]types.Object, len(s)*2)
		i := 0
		for k, v := range s {
			var value types.Object
			if err := bsonToPrimitive(v, &value); err != nil {
				return err
			}
			pairs[i*2] = types.NewString(externalKey(k))
			pairs[i*2+1] = value
			i += 1
		}
		*v = types.NewMap(pairs...)
		return nil
	} else if s, err := types.MarshalBinary(data); err == nil {
		*v = s
		return nil
	}
	return errors.WithStack(encoding.ErrInvalidArgument)
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
