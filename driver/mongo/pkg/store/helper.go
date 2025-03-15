package store

import (
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/types"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func toBSON(val types.Value) (any, error) {
	switch v := val.(type) {
	case types.Map:
		bsonMap := make(map[string]any, v.Len())
		for k, v := range v.Range() {
			key, ok := k.(types.String)
			if !ok {
				return nil, errors.WithStack(encoding.ErrUnsupportedType)
			}

			if key.String() == "id" {
				key = types.NewString("_id")
			}

			val, err := toBSON(v)
			if err != nil {
				return nil, err
			}

			bsonMap[key.String()] = val
		}
		return bsonMap, nil

	case types.Slice:
		bsonSlice := make([]any, v.Len())
		for i, item := range v.Range() {
			val, err := toBSON(item)
			if err != nil {
				return nil, err
			}
			bsonSlice[i] = val
		}
		return bsonSlice, nil

	case types.Binary:
		return bson.Binary{Data: v.Bytes()}, nil

	default:
		return types.InterfaceOf(val), nil
	}
}

func fromBSON(val any) (types.Value, error) {
	switch v := val.(type) {
	case bson.M:
		pairs := make([]types.Value, 0, len(v)*2)
		for k, v := range v {
			key := k
			if k == "_id" {
				key = "id"
			}

			value, err := fromBSON(v)
			if err != nil {
				return nil, err
			}

			pairs = append(pairs, types.NewString(key), value)
		}
		return types.NewMap(pairs...), nil

	case bson.D:
		pairs := make([]types.Value, 0, len(v)*2)
		for _, elem := range v {
			key := elem.Key
			if key == "_id" {
				key = "id"
			}

			value, err := fromBSON(elem.Value)
			if err != nil {
				return nil, err
			}

			pairs = append(pairs, types.NewString(key), value)
		}
		return types.NewMap(pairs...), nil

	case bson.A:
		elements := make([]types.Value, len(v))
		for i, item := range v {
			value, err := fromBSON(item)
			if err != nil {
				return nil, err
			}
			elements[i] = value
		}
		return types.NewSlice(elements...), nil

	case bson.E:
		key := v.Key
		if key == "_id" {
			key = "id"
		}

		value, err := fromBSON(v.Value)
		if err != nil {
			return nil, err
		}
		return types.NewMap(types.NewString(v.Key), value), nil

	case bson.Binary:
		return types.NewBinary(v.Data), nil

	case bson.Null, bson.Undefined, nil:
		return nil, nil

	default:
		return types.Marshal(val)
	}
}
