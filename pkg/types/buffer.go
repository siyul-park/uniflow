package types

import (
	"bytes"
	"encoding"
	"io"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"

	encoding2 "github.com/siyul-park/uniflow/pkg/encoding"
)

// Buffer is a representation of a io.Reader value.
type Buffer = *_buffer

type _buffer struct {
	value io.Reader
}

var (
	_ Value                      = (Buffer)(nil)
	_ io.Reader                  = (Buffer)(nil)
	_ encoding.BinaryMarshaler   = (Buffer)(nil)
	_ encoding.BinaryUnmarshaler = (Buffer)(nil)
)

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
	data, err := io.ReadAll(b.value)
	if err != nil {
		return nil, err
	}
	if err := b.Close(); err != nil {
		return nil, err
	}
	return data, nil
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

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (b Buffer) MarshalBinary() ([]byte, error) {
	return b.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (b Buffer) UnmarshalBinary(data []byte) error {
	if err := b.Close(); err != nil {
		return err
	}

	b.value = bytes.NewBuffer(data)
	return nil
}

func newBufferEncoder() encoding2.EncodeCompiler[any, Value] {
	typeReader := reflect.TypeOf((*io.Reader)(nil)).Elem()

	return encoding2.EncodeCompilerFunc[any, Value](func(typ reflect.Type) (encoding2.Encoder[any, Value], error) {
		if typ == nil {
			return nil, errors.WithStack(encoding2.ErrUnsupportedType)
		} else if typ.ConvertibleTo(typeReader) {
			return encoding2.EncodeFunc(func(source any) (Value, error) {
				s := source.(io.Reader)
				return NewBuffer(s), nil
			}), nil
		}
		return nil, errors.WithStack(encoding2.ErrUnsupportedType)
	})
}

func newBufferDecoder() encoding2.DecodeCompiler[Value] {
	typeReader := reflect.TypeOf((*io.Reader)(nil)).Elem()
	typeBinaryUnmarshaler := reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()
	typeTextUnmarshaler := reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

	return encoding2.DecodeCompilerFunc[Value](func(typ reflect.Type) (encoding2.Decoder[Value, unsafe.Pointer], error) {
		if typ == nil {
			return nil, errors.WithStack(encoding2.ErrUnsupportedType)
		} else if typ.ConvertibleTo(typeBinaryUnmarshaler) {
			return encoding2.DecodeFunc(func(source Value, target unsafe.Pointer) error {
				if s, ok := source.(Buffer); ok {
					t := reflect.NewAt(typ.Elem(), target).Interface().(encoding.BinaryUnmarshaler)
					if data, err := s.Bytes(); err != nil {
						return errors.Wrap(encoding2.ErrUnsupportedValue, err.Error())
					} else if err := t.UnmarshalBinary(data); err != nil {
						return errors.Wrap(encoding2.ErrUnsupportedValue, err.Error())
					}
					return nil
				}
				return errors.WithStack(encoding2.ErrUnsupportedType)
			}), nil
		} else if typ.ConvertibleTo(typeTextUnmarshaler) {
			return encoding2.DecodeFunc(func(source Value, target unsafe.Pointer) error {
				if s, ok := source.(Buffer); ok {
					t := reflect.NewAt(typ.Elem(), target).Interface().(encoding.TextUnmarshaler)
					if data, err := s.Bytes(); err != nil {
						return errors.Wrap(encoding2.ErrUnsupportedValue, err.Error())
					} else if err := t.UnmarshalText(data); err != nil {
						return errors.Wrap(encoding2.ErrUnsupportedValue, err.Error())
					}
					return nil
				}
				return errors.WithStack(encoding2.ErrUnsupportedType)
			}), nil
		} else if typ.Kind() == reflect.Pointer {
			if typ.Elem().ConvertibleTo(typeReader) {
				return encoding2.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Buffer); ok {
						t := reflect.NewAt(typ.Elem(), target)
						t.Elem().Set(reflect.ValueOf(s.Interface()))
						return nil
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.Slice && typ.Elem().Elem().Kind() == reflect.Uint8 {
				return encoding2.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Buffer); ok {
						data, err := s.Bytes()
						if err != nil {
							return err
						}
						t := reflect.NewAt(typ.Elem(), target).Elem()
						t.Set(reflect.AppendSlice(t, reflect.ValueOf(data).Convert(t.Type())))
						return nil
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.Array && typ.Elem().Elem().Kind() == reflect.Uint8 {
				return encoding2.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Buffer); ok {
						data, err := s.Bytes()
						if err != nil {
							return err
						}
						t := reflect.NewAt(typ.Elem(), target).Elem()
						reflect.Copy(t, reflect.ValueOf(data).Convert(t.Type()))
						return nil
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.String {
				return encoding2.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Buffer); ok {
						data, err := s.Bytes()
						if err != nil {
							return err
						}
						*(*string)(target) = string(data)
						return nil
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem() == types[KindUnknown] {
				return encoding2.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Buffer); ok {
						*(*any)(target) = s.Interface()
						return nil
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			}
		}
		return nil, errors.WithStack(encoding2.ErrUnsupportedType)
	})
}
