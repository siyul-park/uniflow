package runtime

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/stretchr/testify/require"
)

func TestCatalog_Table(t *testing.T) {
	agent := runtime.NewAgent()
	catalog := NewCatalog(agent)

	tests := []string{"frames", "processes", "symbols"}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			table, err := catalog.Table(tt)
			require.NoError(t, err)
			require.NotNil(t, table)
		})
	}
}
