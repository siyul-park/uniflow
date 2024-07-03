package mime

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompress(t *testing.T) {
	tests := []struct {
		name     string
		encoding string
	}{
		{name: "Gzip", encoding: EncodingGzip},
		{name: "Deflate", encoding: EncodingDeflate},
		{name: "Brotli", encoding: EncodingBr},
		{name: "Identity", encoding: EncodingIdentity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var encoded bytes.Buffer
			writer, err := Compress(&encoded, tt.encoding)
			assert.NoError(t, err)

			data := []byte("hello, world")
			_, err = writer.Write(data)
			assert.NoError(t, err)

			if closer, ok := writer.(io.Closer); ok {
				assert.NoError(t, closer.Close())
			}

			var decoded bytes.Buffer
			reader, err := Decompress(&encoded, tt.encoding)
			assert.NoError(t, err)

			_, err = io.Copy(&decoded, reader)
			assert.NoError(t, err)
			assert.Equal(t, data, decoded.Bytes())
		})
	}
}
