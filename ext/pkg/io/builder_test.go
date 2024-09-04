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

	tests := []string{KindSQL, KindWrite}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			_, ok := s.KnownType(tt)
			assert.True(t, ok)

			_, ok = s.Codec(tt)
			assert.True(t, ok)
		})
	}
}
