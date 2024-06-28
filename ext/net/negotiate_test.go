package net

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestNegotiate(t *testing.T) {
	testCases := []struct {
		when   string
		offers []string
		expect string
	}{

		{
			when:   "",
			offers: []string{faker.Word(), faker.Word(), faker.Word()},
			expect: "",
		},
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
		{
			when:   "text/*, application/json;q=1.0, *;q=0.5",
			offers: []string{"text/plain"},
			expect: "text/plain",
		},
		{
			when:   "*",
			offers: []string{"text/plain"},
			expect: "text/plain",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expect, func(t *testing.T) {
			typ := Negotiate(tc.when, tc.offers)
			assert.Equal(t, tc.expect, typ)
		})
	}
}
