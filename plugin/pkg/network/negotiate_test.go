package network

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestNegotiate(t *testing.T) {
	testCases := []struct {
		when   string
		expect string
	}{
		{
			when:   "gzip",
			expect: "gzip",
		},
		{
			when:   "gzip, compress, br",
			expect: "br",
		},
		{
			when:   "deflate, gzip;q=1.0, *;q=0.5",
			expect: "gzip",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expect, func(t *testing.T) {
			typ := Negotiate(tc.when, nil)
			assert.Equal(t, tc.expect, typ)
		})
	}
}

func TestCompressAndDecompress(t *testing.T) {
	testCases := []struct {
		data     []byte
		encoding string
	}{
		{
			data:     []byte(faker.Sentence()),
			encoding: EncodingGzip,
		},
		{
			data:     []byte(faker.Sentence()),
			encoding: EncodingDeflate,
		},
		{
			data:     []byte(faker.Sentence()),
			encoding: EncodingBr,
		},
		{
			data:     []byte(faker.Sentence()),
			encoding: EncodingIdentity,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.encoding, func(t *testing.T) {
			b, err := Compress(tc.data, tc.encoding)
			assert.NoError(t, err)

			b, err = Decompress(b, tc.encoding)
			assert.NoError(t, err)
			assert.Equal(t, tc.data, b)
		})
	}
}
