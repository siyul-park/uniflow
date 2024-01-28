package network

import (
	"testing"

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
			expect: "gzip",
		},
		{
			when:   "deflate, gzip;q=1.0, *;q=0.5",
			expect: "deflate",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expect, func(t *testing.T) {
			typ := Negotiate(tc.when, nil)
			assert.Equal(t, tc.expect, typ)
		})
	}
}
