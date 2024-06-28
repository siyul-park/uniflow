package net

import (
	"bytes"
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
func Compress(data []byte, encoding string) ([]byte, error) {
	var b bytes.Buffer
	var w io.Writer
	switch encoding {
	case EncodingGzip:
		w = gzip.NewWriter(&b)
	case EncodingDeflate:
		w = zlib.NewWriter(&b)
	case EncodingBr:
		w = brotli.NewWriter(&b)
	default:
		return data, nil
	}

	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if w, ok := w.(io.Closer); ok {
		w.Close()
	}

	return b.Bytes(), nil
}

// Decompress decompresses input data using the specified encoding, returns original if unsupported.
func Decompress(data []byte, encoding string) ([]byte, error) {
	var r io.Reader
	var err error
	switch encoding {
	case EncodingGzip:
		r, err = gzip.NewReader(bytes.NewReader(data))
	case EncodingDeflate:
		r, err = zlib.NewReader(bytes.NewReader(data))
	case EncodingBr:
		r = brotli.NewReader(bytes.NewReader(data))
	default:
		return data, nil
	}
	if err != nil {
		return nil, err
	}
	if r, ok := r.(io.Closer); ok {
		defer r.Close()
	}

	return io.ReadAll(r)
}
