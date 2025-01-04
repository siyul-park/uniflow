package mime

import (
	"compress/gzip"
	"compress/zlib"
	"io"

	"github.com/andybalholm/brotli"
)

type multiWriter struct {
	pipe []io.Writer
}

type multiReader struct {
	pipe []io.Reader
}

const (
	EncodingGzip     = "gzip"
	EncodingDeflate  = "deflate"
	EncodingBr       = "br"
	EncodingIdentity = "identity"
)

var _ io.Writer = (*multiWriter)(nil)
var _ io.Closer = (*multiWriter)(nil)
var _ io.Reader = (*multiReader)(nil)
var _ io.Closer = (*multiReader)(nil)

// Compress compresses input data using the specified encoding, returns original if unsupported.
func Compress(writer io.Writer, encoding string) (io.Writer, error) {
	switch encoding {
	case EncodingGzip:
		return newMultiWriter(writer, gzip.NewWriter(writer)), nil
	case EncodingDeflate:
		return newMultiWriter(writer, zlib.NewWriter(writer)), nil
	case EncodingBr:
		return newMultiWriter(writer, brotli.NewWriter(writer)), nil
	default:
		return writer, nil
	}
}

// Decompress decompresses input data using the specified encoding, returns original if unsupported.
func Decompress(reader io.Reader, encoding string) (io.Reader, error) {
	switch encoding {
	case EncodingGzip:
		r, err := gzip.NewReader(reader)
		if err != nil {
			return nil, err
		}
		return newMultiReader(reader, r), nil
	case EncodingDeflate:
		r, err := zlib.NewReader(reader)
		if err != nil {
			return nil, err
		}
		return newMultiReader(reader, r), nil
	case EncodingBr:
		return newMultiReader(reader, brotli.NewReader(reader)), nil
	default:
		return reader, nil
	}
}

// newMultiWriter creates a writer that writes to multiple writers in sequence.
func newMultiWriter(pipe ...io.Writer) io.Writer {
	return &multiWriter{pipe: pipe}
}

// newMultiReader creates a reader that reads from multiple readers in sequence.
func newMultiReader(pipe ...io.Reader) io.Reader {
	return &multiReader{pipe: pipe}
}

func (w *multiWriter) Write(p []byte) (n int, err error) {
	if len(w.pipe) == 0 {
		return 0, io.ErrClosedPipe
	}
	return w.pipe[len(w.pipe)-1].Write(p)
}

func (w *multiWriter) Close() error {
	for i := len(w.pipe) - 1; i >= 0; i-- {
		if c, ok := w.pipe[i].(io.Closer); ok {
			if err := c.Close(); err != nil {
				return err
			}
		}
	}
	w.pipe = nil
	return nil
}

func (r *multiReader) Read(p []byte) (n int, err error) {
	if len(r.pipe) == 0 {
		return 0, io.ErrClosedPipe
	}
	return r.pipe[len(r.pipe)-1].Read(p)
}

func (r *multiReader) Close() error {
	for i := len(r.pipe) - 1; i >= 0; i-- {
		if c, ok := r.pipe[i].(io.Closer); ok {
			if err := c.Close(); err != nil {
				return err
			}
		}
	}
	r.pipe = nil
	return nil
}
