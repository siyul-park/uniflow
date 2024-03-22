package primitive

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"io"
	"reflect"
	"sync"
)

// Buffer is a struct representing a buffer with a reader and content.
type Buffer struct {
	reader  io.Reader
	content []byte
	mu      sync.RWMutex
}

type bufferReader struct {
	buffer *Buffer
	offset int
}

var _ Value = (*Buffer)(nil)
var _ io.Reader = (*bufferReader)(nil)

// NewBuffer creates a new Buffer instance with the given io.Reader.
func NewBuffer(value io.Reader) *Buffer {
	return &Buffer{reader: value}
}

// Read reads a specific size from the buffer content.
func (b *Buffer) Read(size int) ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.read(size)
}

// Bytes returns the entire content of the buffer.
func (b *Buffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()

	data, _ := b.readAll()
	return data
}

// Reader returns an io.Reader for reading from the buffer.
func (b *Buffer) Reader() io.Reader {
	b.mu.Lock()
	defer b.mu.Unlock()

	return &bufferReader{buffer: b}
}

// Kind returns the kind of the value, which is KindBuffer.
func (b *Buffer) Kind() Kind {
	return KindBuffer
}

// Compare compares two Buffer instances by comparing their content.
func (b *Buffer) Compare(v Value) int {
	if other, ok := v.(*Buffer); ok {
		return bytes.Compare(b.Bytes(), other.Bytes())
	}
	if b.Kind() > v.Kind() {
		return 1
	}
	return -1
}

// Interface returns the buffer's reader as an interface{}.
func (b *Buffer) Interface() any {
	return b.Reader()
}

func (b *Buffer) readAll() ([]byte, error) {
	for i := 1; ; i++ {
		if data, err := b.read(512 * i); errors.Is(err, io.EOF) {
			return data, nil
		} else if err != nil {
			return nil, err
		}
	}
}

func (b *Buffer) read(size int) ([]byte, error) {
	var err error
	for len(b.content) < size {
		buffer := make([]byte, 0, 512)
		for {
			var n int
			n, err = b.reader.Read(buffer[len(buffer):cap(buffer)])
			buffer = buffer[:len(buffer)+n]
			if err != nil {
				break
			}

			if len(buffer) == cap(buffer) {
				buffer = append(buffer, 0)[:len(buffer)]
			}
		}

		b.content = append(b.content, buffer...)
		if err != nil {
			break
		}
	}
	return b.content, err
}

// Read reads data from the bufferReader.
func (b *bufferReader) Read(p []byte) (int, error) {
	data, err := b.buffer.Read(b.offset + cap(p))
	if err != nil && !errors.Is(err, io.EOF) {
		return 0, err
	}

	n := copy(p, data[b.offset:])
	b.offset += n

	return n, err
}

func newBufferEncoder() encoding.Encoder[any, Value] {
	return encoding.EncoderFunc[any, Value](func(source any) (Value, error) {
		if s, ok := source.(io.Reader); ok {
			return NewBuffer(s), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newBufferDecoder() encoding.Decoder[Value, any] {
	typeReader := reflect.TypeOf((*io.Reader)(nil)).Elem()
	binaryDecoder := newBinaryDecoder()

	return encoding.DecoderFunc[Value, any](func(source Value, target any) error {
		if s, ok := source.(*Buffer); ok {
			if t := reflect.ValueOf(target); t.Kind() == reflect.Pointer {
				if t.Elem().Type() == typeReader {
					t.Elem().Set(reflect.ValueOf(s.Interface()))
					return nil
				} else if t.Elem().Type() == typeAny {
					t.Elem().Set(reflect.ValueOf(s.Interface()))
					return nil
				}
			}
			return binaryDecoder.Decode(NewBinary(s.Bytes()), target)
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
