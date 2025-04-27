package driver

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/plugins/mongodb/internal/server"
)

func TestDriver_Open(t *testing.T) {
	d := New()
	defer d.Close()

	srv := server.New()
	defer server.Release(srv)

	c, err := d.Open(srv.URI())
	require.NoError(t, err)
	require.NotNil(t, c)
}
