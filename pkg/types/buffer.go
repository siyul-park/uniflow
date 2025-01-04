package types

import (
	"encoding/base64"
	"io"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

// Buffer is a representation of a io.Reader value.
type Buffer = *_buffer

type _buffer struct {
	value io.Reader
}

var _ Value = (Buffer)(nil)
var _ io.Reader = (Buffer)(nil)

// NewBuffer creates a new Buffer instance.
func NewBuffer(value io.Reader) Buffer {
	return &_buffer{value: value}
}

// Read reads data from the buffer into p.
func (b Buffer) Read(p []byte) (n int, err error) {
	return b.value.Read(p)
}

// Bytes returns the raw byte slice.
func (b Buffer) Bytes() ([]byte, error) {
	bytes, err := io.ReadAll(b.value)
	if err != nil {
		return nil, err
	}
	closer, ok := b.value.(io.Closer)
	if ok {
		if err := closer.Close(); err != nil {
			return nil, err
		}
	}
	return bytes, nil
}

// Close closes the buffer.
func (b Buffer) Close() error {
	if closer, ok := b.value.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// Kind returns the kind of the buffer.
func (b Buffer) Kind() Kind {
	return KindBuffer
}

// Hash returns a hash value for the buffer.
func (b Buffer) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(b)))
}

// Interface returns the underlying io.Reader.
func (b Buffer) Interface() any {
	return b.value
}

// Equal checks if the buffer is equal to another Value.
func (b Buffer) Equal(other Value) bool {
	if o, ok := other.(Buffer); ok {
		return b == o
	}
	return false
}

// Compare compares the buffer with another Value.
func (b Buffer) Compare(other Value) int {
	if o, ok := other.(Buffer); ok {
		return compare(b.Hash(), o.Hash())
	}
	return compare(b.Kind(), KindOf(other))
}

func newBufferEncoder() encoding.EncodeCompiler[any, Value] {
	typeReader := reflect.TypeOf((*io.Reader)(nil)).Elem()

	return encoding.EncodeCompilerFunc[any, Value](func(typ reflect.Type) (encoding.Encoder[any, Value], error) {
		if typ == nil {
			return nil, errors.WithStack(encoding.ErrUnsupportedType)
		} else if typ.ConvertibleTo(typeReader) {
			return encoding.EncodeFunc(func(source any) (Value, error) {
				s := source.(io.Reader)
				return NewBuffer(s), nil
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}

func newBufferDecoder() encoding.DecodeCompiler[Value] {
	typeReader := reflect.TypeOf((*io.Reader)(nil)).Elem()

	return encoding.DecodeCompilerFunc[Value](func(typ reflect.Type) (encoding.Decoder[Value, unsafe.Pointer], error) {
		if typ == nil {
			return nil, errors.WithStack(encoding.ErrUnsupportedType)
		} else if typ.Kind() == reflect.Pointer {
			if typ.Elem().ConvertibleTo(typeReader) {
				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Buffer); ok {
						t := reflect.NewAt(typ.Elem(), target)
						t.Elem().Set(reflect.ValueOf(s.Interface()))
						return nil
					}
					return errors.WithStack(encoding.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.Slice && typ.Elem().Elem().Kind() == reflect.Uint8 {
				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Buffer); ok {
						bytes, err := s.Bytes()
						if err != nil {
							return err
						}
						t := reflect.NewAt(typ.Elem(), target).Elem()
						t.Set(reflect.AppendSlice(t, reflect.ValueOf(bytes).Convert(t.Type())))
						return nil
					}
					return errors.WithStack(encoding.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.Array && typ.Elem().Elem().Kind() == reflect.Uint8 {
				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Buffer); ok {
						bytes, err := s.Bytes()
						if err != nil {
							return err
						}
						t := reflect.NewAt(typ.Elem(), target).Elem()
						reflect.Copy(t, reflect.ValueOf(bytes).Convert(t.Type()))
						return nil
					}
					return errors.WithStack(encoding.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.String {
				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Buffer); ok {
						bytes, err := io.ReadAll(s)
						if err != nil {
							return err
						}
						*(*string)(target) = base64.StdEncoding.EncodeToString(bytes)
						return nil
					}
					return errors.WithStack(encoding.ErrUnsupportedType)
				}), nil
			} else if typ.Elem() == types[KindUnknown] {
				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Buffer); ok {
						*(*any)(target) = s.Interface()
						return nil
					}
					return errors.WithStack(encoding.ErrUnsupportedType)
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}
