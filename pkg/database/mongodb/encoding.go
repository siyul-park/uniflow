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
	documentEncoder = NewDocumentEncoder()
	documentDecoder = NewDocumentDecoder()

	filterEncoder = NewFilterEncoder(documentEncoder)
	filterDecoder = NewFilterDecoder(documentDecoder)
)

var (
	toLowerCamel = changeCase(strcase.ToLowerCamel)
	toSnake      = changeCase(strcase.ToSnake)
)

func MarshalFilter(v *database.Filter) (any, error) {
	return filterEncoder.Encode(v)
}

func UnmarshalFilter(data any, v **database.Filter) error {
	return filterDecoder.Decode(data, v)
}

func MarshalDocument(v primitive.Object) (any, error) {
	return documentEncoder.Encode(v)
}

func UnmarshalDocument(data any, v *primitive.Object) error {
	return documentDecoder.Decode(data, v)
}

func NewFilterEncoder(encoder encoding.Encoder[primitive.Object, any]) encoding.Encoder[*database.Filter, any] {
	return encoding.EncoderFunc[*database.Filter, any](func(source *database.Filter) (any, error) {
		if source == nil {
			return bson.D{}, nil
		}

		self := NewFilterEncoder(encoder)

		switch source.OP {
		case database.AND, database.OR:
			var values bson.A
			for _, e := range source.Children {
				if value, err := self.Encode(e); err != nil {
					return nil, err
				} else {
					values = append(values, value)
				}
			}

			if source.OP == database.AND {
				return bson.D{{Key: "$and", Value: values}}, nil
			} else if source.OP == database.OR {
				return bson.D{{Key: "$or", Value: values}}, nil
			}
		case database.NULL, database.NNULL:
			k := bsonKey(source.Key)

			if source.OP == database.NULL {
				return bson.D{{Key: k, Value: bson.M{"$eq": nil}}}, nil
			} else if source.OP == database.NNULL {
				return bson.D{{Key: k, Value: bson.M{"$ne": nil}}}, nil
			}
		default:
			k := bsonKey(source.Key)
			v, err := encoder.Encode(source.Value)
			if err != nil {
				return nil, err
			}

			if source.OP == database.EQ {
				return bson.D{{Key: k, Value: bson.M{"$eq": v}}}, nil
			} else if source.OP == database.NE {
				return bson.D{{Key: k, Value: bson.M{"$ne": v}}}, nil
			} else if source.OP == database.LT {
				return bson.D{{Key: k, Value: bson.M{"$lt": v}}}, nil
			} else if source.OP == database.LTE {
				return bson.D{{Key: k, Value: bson.M{"$lte": v}}}, nil
			} else if source.OP == database.GT {
				return bson.D{{Key: k, Value: bson.M{"$gt": v}}}, nil
			} else if source.OP == database.GTE {
				return bson.D{{Key: k, Value: bson.M{"$gte": v}}}, nil
			} else if source.OP == database.IN {
				return bson.D{{Key: k, Value: bson.M{"$in": v}}}, nil
			} else if source.OP == database.NIN {
				return bson.D{{Key: k, Value: bson.M{"$nin": v}}}, nil
			}
		}

		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func NewFilterDecoder(decoder encoding.Decoder[any, *primitive.Object]) encoding.Decoder[any, **database.Filter] {
	return encoding.DecoderFunc[any, **database.Filter](func(source any, target **database.Filter) error {
		s, ok := bsonMA(source)
		if !ok {
			return errors.WithStack(encoding.ErrUnsupportedValue)
		}

		self := NewFilterDecoder(decoder)

		var children []*database.Filter
		for _, curr := range s {
			for key, value := range curr {
				if key == "$and" || key == "$or" {
					if value, ok := bsonMA(value); !ok {
						return errors.WithStack(encoding.ErrUnsupportedValue)
					} else {
						var values []*database.Filter
						for _, v := range value {
							var value *database.Filter
							if err := self.Decode(v, &value); err != nil {
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
					if err := self.Decode(value, &child); err != nil {
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
						return errors.WithStack(encoding.ErrUnsupportedValue)
					}
					children = append(children, child)
				} else if value, ok := bsonM(value); ok {
					for op, v := range value {
						if !strings.HasPrefix(op, "$") {
							return errors.WithStack(encoding.ErrUnsupportedValue)
						}
						child := &database.Filter{
							Key: key,
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
							return errors.WithStack(encoding.ErrUnsupportedValue)
						}

						var value primitive.Object
						if err := decoder.Decode(v, &value); err != nil {
							return err
						}
						child.Value = value
						children = append(children, child)
					}
				} else {
					return errors.WithStack(encoding.ErrUnsupportedValue)
				}
			}
		}

		if len(children) == 0 {
			*target = nil
		} else if len(children) == 1 {
			*target = children[0]
		} else {
			*target = &database.Filter{
				OP:       database.AND,
				Children: children,
			}
		}

		return nil
	})
}

func NewDocumentEncoder() encoding.Encoder[primitive.Object, any] {
	return encoding.EncoderFunc[primitive.Object, any](func(source primitive.Object) (any, error) {
		if source == nil {
			return bsonprimitive.Null{}, nil
		}

		self := NewDocumentEncoder()

		if s, ok := source.(primitive.Binary); ok {
			return bsonprimitive.Binary{Data: []byte(s)}, nil
		} else if s, ok := source.(*primitive.Map); ok {
			t := make(bsonprimitive.M, s.Len())
			for _, k := range s.Keys() {
				v, _ := s.Get(k)
				if k, ok := k.(primitive.String); !ok {
					return nil, errors.WithStack(encoding.ErrUnsupportedValue)
				} else {
					if v, err := self.Encode(v); err != nil {
						return nil, err
					} else {
						t[bsonKey(k.String())] = v
					}
				}
			}
			return t, nil
		} else if s, ok := source.(*primitive.Slice); ok {
			t := make(bsonprimitive.A, s.Len())
			for i := 0; i < s.Len(); i++ {
				if v, err := self.Encode(s.Get(i)); err != nil {
					return nil, err
				} else {
					t[i] = v
				}
			}
			return t, nil
		} else {
			return source.Interface(), nil
		}
	})
}

func NewDocumentDecoder() encoding.Decoder[any, *primitive.Object] {
	return encoding.DecoderFunc[any, *primitive.Object](func(source any, target *primitive.Object) error {
		self := NewDocumentDecoder()

		if source == nil {
			*target = nil
			return nil
		} else if _, ok := source.(bsonprimitive.Null); ok {
			*target = nil
			return nil
		} else if _, ok := source.(bsonprimitive.Undefined); ok {
			*target = nil
			return nil
		} else if s, ok := source.(bsonprimitive.Binary); ok {
			*target = primitive.NewBinary(s.Data)
			return nil
		} else if s, ok := source.(bsonprimitive.A); ok {
			values := make([]primitive.Object, len(s))
			for i, e := range s {
				var value primitive.Object
				if err := self.Decode(e, &value); err != nil {
					return err
				}
				values[i] = value
			}
			*target = primitive.NewSlice(values...)
			return nil
		} else if s, ok := source.(bsonprimitive.D); ok {
			pairs := make([]primitive.Object, len(s)*2)
			for i, e := range s {
				var value primitive.Object
				if err := self.Decode(e.Value, &value); err != nil {
					return err
				}
				pairs[i*2] = primitive.NewString(documentKey(e.Key))
				pairs[i*2+1] = value
			}
			*target = primitive.NewMap(pairs...)
			return nil
		} else if s, ok := source.(bsonprimitive.M); ok {
			pairs := make([]primitive.Object, len(s)*2)
			i := 0
			for k, v := range s {
				var value primitive.Object
				if err := self.Decode(v, &value); err != nil {
					return err
				}
				pairs[i*2] = primitive.NewString(documentKey(k))
				pairs[i*2+1] = value
				i += 1
			}
			*target = primitive.NewMap(pairs...)
			return nil
		} else if s, err := primitive.MarshalBinary(source); err == nil {
			*target = s
			return nil
		}

		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func mongoSorts(sorts []database.Sort) bson.D {
	sort := bson.D{}
	for _, s := range sorts {
		sort = append(sort, bson.E{
			Key:   bsonKey(s.Key),
			Value: mongoOrder(s.Order),
		})
	}
	return sort
}

func mongoOrder(order database.Order) int {
	if order == database.OrderASC {
		return 1
	}
	return -1
}

func bsonKey(key string) string {
	if key == "id" {
		return "_id"
	}
	return toLowerCamel(key)
}

func documentKey(key string) string {
	if key == "_id" {
		return "id"
	}
	return toSnake(key)
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

func bsonMA(value any) ([]bson.M, bool) {
	if m, ok := bsonM(value); ok {
		return []bson.M{m}, true
	}

	var m []bson.M
	if v, ok := value.([]bson.M); ok {
		m = v
	} else if v, ok := value.([]bson.D); ok {
		for _, e := range v {
			if e, ok := bsonM(e); ok {
				m = append(m, e)
			} else {
				return nil, false
			}
		}
	} else if v, ok := value.([]any); ok {
		for _, e := range v {
			if e, ok := bsonM(e); ok {
				m = append(m, e)
			} else {
				return nil, false
			}
		}
	} else {
		return nil, false
	}

	return m, true
}

func bsonM(value any) (bson.M, bool) {
	var m bson.M
	if v, ok := value.(bson.M); ok {
		m = v
	} else if v, ok := value.(bson.D); ok {
		m := make(bson.M, len(v))
		for _, e := range v {
			m[e.Key] = e.Value
		}
		return m, true
	} else {
		return nil, false
	}

	return m, true
}
