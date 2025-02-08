package types

import (
	"encoding/json"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

func newJSONEncoder(encoder *encoding.EncodeAssembler[any, Value]) encoding.EncodeCompiler[any, Value] {
	typeJSONMarshaler := reflect.TypeOf((*json.Marshaler)(nil)).Elem()

	return encoding.EncodeCompilerFunc[any, Value](func(typ reflect.Type) (encoding.Encoder[any, Value], error) {
		if typ != nil && typ.ConvertibleTo(typeJSONMarshaler) {
			return encoding.EncodeFunc(func(source any) (Value, error) {
				s := source.(json.Marshaler)
				data, err := s.MarshalJSON()
				if err != nil {
					return nil, errors.Wrap(encoding.ErrUnsupportedValue, err.Error())
				}
				var val any
				if err := json.Unmarshal(data, &val); err != nil {
					return nil, errors.Wrap(encoding.ErrUnsupportedValue, err.Error())
				}
				return encoder.Encode(val)
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}

func newJSONDecoder(decoder *encoding.DecodeAssembler[Value, any]) encoding.DecodeCompiler[Value] {
	typeJSONUnmarshaler := reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()

	return encoding.DecodeCompilerFunc[Value](func(typ reflect.Type) (encoding.Decoder[Value, unsafe.Pointer], error) {
		if typ != nil && typ.Kind() == reflect.Pointer {
			if typ.ConvertibleTo(typeJSONUnmarshaler) {
				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					t := reflect.NewAt(typ.Elem(), target).Interface().(json.Unmarshaler)
					var val any
					if err := decoder.Decode(source, &val); err != nil {
						return err
					}
					data, err := json.Marshal(val)
					if err != nil {
						return errors.Wrap(encoding.ErrUnsupportedValue, err.Error())
					}
					if err := t.UnmarshalJSON(data); err != nil {
						return errors.Wrap(encoding.ErrUnsupportedValue, err.Error())
					}
					return nil
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}
