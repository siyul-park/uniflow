package types

import (
	"reflect"
	"time"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

func newTimeEncoder() encoding.EncodeCompiler[any, Value] {
	typeTime := reflect.TypeOf((*time.Time)(nil)).Elem()

	return encoding.EncodeCompilerFunc[any, Value](func(typ reflect.Type) (encoding.Encoder[any, Value], error) {
		if typ != nil && typ == typeTime {
			return encoding.EncodeFunc(func(source any) (Value, error) {
				s := source.(time.Time)
				return NewInt64(s.UnixMilli()), nil
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}

func newTimeDecoder() encoding.DecodeCompiler[Value] {
	typeTime := reflect.TypeOf((*time.Time)(nil)).Elem()

	return encoding.DecodeCompilerFunc[Value](func(typ reflect.Type) (encoding.Decoder[Value, unsafe.Pointer], error) {
		if typ != nil && typ.Kind() == reflect.Pointer {
			if typ.Elem() == typeTime {
				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					var v time.Time
					var err error
					if s, ok := source.(String); ok {
						v, err = time.Parse(time.RFC3339, s.String())
					} else if s, ok := source.(Integer); ok {
						v = time.UnixMilli(s.Int()).UTC()
					} else if s, ok := source.(Float); ok {
						v = time.UnixMilli(int64(s.Float())).UTC()
					} else {
						err = errors.WithStack(encoding.ErrUnsupportedType)
					}
					if err != nil {
						return err
					}
					t := reflect.NewAt(typ.Elem(), target)
					t.Elem().Set(reflect.ValueOf(v))
					return nil
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}

func newDurationEncoder() encoding.EncodeCompiler[any, Value] {
	typeDuration := reflect.TypeOf((*time.Duration)(nil)).Elem()

	return encoding.EncodeCompilerFunc[any, Value](func(typ reflect.Type) (encoding.Encoder[any, Value], error) {
		if typ != nil && typ == typeDuration {
			return encoding.EncodeFunc(func(source any) (Value, error) {
				s := source.(time.Duration)
				return NewInt64(s.Milliseconds()), nil
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}

func newDurationDecoder() encoding.DecodeCompiler[Value] {
	typeDuration := reflect.TypeOf((*time.Duration)(nil)).Elem()

	return encoding.DecodeCompilerFunc[Value](func(typ reflect.Type) (encoding.Decoder[Value, unsafe.Pointer], error) {
		if typ != nil && typ.Kind() == reflect.Pointer {
			if typ.Elem() == typeDuration {
				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					var v time.Duration
					var err error
					if s, ok := source.(String); ok {
						if v, err = time.ParseDuration(s.String()); err != nil {
							err = errors.WithMessage(encoding.ErrUnsupportedValue, err.Error())
						}
					} else if s, ok := source.(Integer); ok {
						v = time.Millisecond * (time.Duration)(s.Int())
					} else if s, ok := source.(Float); ok {
						v = time.Millisecond * (time.Duration)(s.Float())
					} else {
						err = errors.WithStack(encoding.ErrUnsupportedType)
					}
					if err != nil {
						return err
					}
					t := reflect.NewAt(typ.Elem(), target)
					t.Elem().Set(reflect.ValueOf(v))
					return nil
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}
