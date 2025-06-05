package spec

import (
	"reflect"
	"unsafe"

	"github.com/pkg/errors"

	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/types"
)

func init() {
	types.Decoder.Add(newSpecDecoder(types.Decoder))
}

func newSpecDecoder(decoder *encoding.DecodeAssembler[types.Value, any]) encoding.DecodeCompiler[types.Value] {
	typeSpec := reflect.TypeOf((*Spec)(nil)).Elem()

	return encoding.DecodeCompilerFunc[types.Value](func(typ reflect.Type) (encoding.Decoder[types.Value, unsafe.Pointer], error) {
		if typ != nil && typ.Kind() == reflect.Pointer {
			if typ.Elem() == typeSpec {
				unstructured := &Unstructured{}
				child, err := decoder.Compile(reflect.TypeOf(&unstructured))
				if err != nil {
					return nil, err
				}

				return encoding.DecodeFunc(func(source types.Value, target unsafe.Pointer) error {
					unstructured := &Unstructured{}
					if err := child.Decode(source, unsafe.Pointer(&unstructured)); err != nil {
						return err
					}
					*(*Spec)(target) = unstructured
					return nil
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}
