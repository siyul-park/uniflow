package mime

import (
	"compress/gzip"
	"compress/zlib"
	"io"

	"github.com/andybalholm/brotli"
)

const (
	EncodingGzip     = "gzip"
	EncodingDeflate  = "deflate"
	EncodingBr       = "br"
	EncodingIdentity = "identity"
)

// Compress compresses input data using the specified encoding, returns original if unsupported.
func Compress(writer io.Writer, encoding string) (io.Writer, error) {
	var w io.Writer
	switch encoding {
	case EncodingGzip:
		w = gzip.NewWriter(writer)
	case EncodingDeflate:
		w = zlib.NewWriter(writer)
	case EncodingBr:
		w = brotli.NewWriter(writer)
	default:
		w = writer
	}
	return w, nil
}

// Decompress decompresses input data using the specified encoding, returns original if unsupported.
func Decompress(reader io.Reader, encoding string) (io.Reader, error) {
	var r io.Reader
	var err error
	switch encoding {
	case EncodingGzip:
		r, err = gzip.NewReader(reader)
	case EncodingDeflate:
		r, err = zlib.NewReader(reader)
	case EncodingBr:
		r = brotli.NewReader(reader)
	default:
		return reader, nil
	}
	return r, err
}
