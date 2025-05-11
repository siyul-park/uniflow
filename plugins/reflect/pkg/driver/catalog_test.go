package driver

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/driver"
	"github.com/stretchr/testify/require"
)

func TestCatalog_Table(t *testing.T) {
	drv := driver.New()
	defer drv.Close()

	conn, err := drv.Open(faker.UUIDHyphenated())
	require.NoError(t, err)

	defer conn.Close()

	catalog := NewCatalog(conn)

	tests := []string{"specs", "values"}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			table, err := catalog.Table(tt)
			require.NoError(t, err)
			require.NotNil(t, table)
		})
	}
}
