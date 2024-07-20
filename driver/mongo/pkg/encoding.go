package mongo

import (
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	types.BinaryEncoder.Add(newBSONEncoder())
	types.TextEncoder.Add(newBSONEncoder())
	types.Decoder.Add(newBSONDecoder())
}

func newBSONEncoder() encoding.EncodeCompiler[any, types.Value] {
	typeBinary := reflect.TypeOf((*primitive.Binary)(nil)).Elem()

	return encoding.EncodeCompilerFunc[any, types.Value](func(typ reflect.Type) (encoding.Encoder[any, types.Value], error) {
		if typ != nil && typ.ConvertibleTo(typeBinary) {
			return encoding.EncodeFunc[any, types.Value](func(source any) (types.Value, error) {
				s := source.(primitive.Binary)
				return types.NewBinary(s.Data), nil
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}

func newBSONDecoder() encoding.DecodeCompiler[types.Value] {
	typeBinary := reflect.TypeOf((*primitive.Binary)(nil)).Elem()

	return encoding.DecodeCompilerFunc[types.Value](func(typ reflect.Type) (encoding.Decoder[types.Value, unsafe.Pointer], error) {
		if typ != nil && typ.Kind() == reflect.Pointer {
			if typ.Elem().ConvertibleTo(typeBinary) {
				return encoding.DecodeFunc[types.Value, unsafe.Pointer](func(source types.Value, target unsafe.Pointer) error {
					if s, ok := source.(types.Binary); ok {
						t := reflect.NewAt(typ.Elem(), target)
						t.Elem().Set(reflect.ValueOf(primitive.Binary{
							Data: s.Bytes(),
						}))
						return nil
					}
					return errors.WithStack(encoding.ErrUnsupportedType)
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}
