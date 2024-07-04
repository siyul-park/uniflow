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
	switch encoding {
	case EncodingGzip:
		return gzip.NewWriter(writer), nil
	case EncodingDeflate:
		return zlib.NewWriter(writer), nil
	case EncodingBr:
		return brotli.NewWriter(writer), nil
	default:
		return writer, nil
	}
}

// Decompress decompresses input data using the specified encoding, returns original if unsupported.
func Decompress(reader io.Reader, encoding string) (io.Reader, error) {
	switch encoding {
	case EncodingGzip:
		return gzip.NewReader(reader)
	case EncodingDeflate:
		return zlib.NewReader(reader)
	case EncodingBr:
		return brotli.NewReader(reader), nil
	default:
		return reader, nil
	}
}
