package network

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"
	"mime"
	"slices"
	"strconv"
	"strings"

	"github.com/andybalholm/brotli"
)

const (
	EncodingGzip     = "gzip"
	EncodingDeflate  = "deflate"
	EncodingBr       = "br"
	EncodingIdentity = "identity"
)

func Negotiate(value string, offers []string) string {
	tokens := strings.Split(value, ",")

	typ := ""
	quality := 0.0
	for _, token := range tokens {
		if mediaType, params, err := mime.ParseMediaType(strings.Trim(token, " ")); err == nil {
			if offers == nil || slices.Contains(offers, mediaType) {
				if q, _ := strconv.ParseFloat(strings.Trim(params["q"], " "), 32); q >= quality {
					typ = mediaType
					quality = q
				}
			}
		}
	}

	return typ
}

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
