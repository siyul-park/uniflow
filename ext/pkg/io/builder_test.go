package io

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/assert"
)

func TestAddToScheme(t *testing.T) {
	s := scheme.New()

	err := AddToScheme(NewOSFileSystem()).AddToScheme(s)
	assert.NoError(t, err)

	tests := []string{KindSQL, KindPrint, KindScan}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			assert.NotNil(t, s.KnownType(tt))
			assert.NotNil(t, s.Codec(tt))
		})
	}
}
