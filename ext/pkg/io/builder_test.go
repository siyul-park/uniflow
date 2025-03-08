package io

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/require"
)

func TestAddToScheme(t *testing.T) {
	s := scheme.New()

	err := AddToScheme(NewOSFileSystem()).AddToScheme(s)
	require.NoError(t, err)

	tests := []string{KindSQL, KindPrint, KindScan}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			require.NotNil(t, s.KnownType(tt))
			require.NotNil(t, s.Codec(tt))
		})
	}
}
