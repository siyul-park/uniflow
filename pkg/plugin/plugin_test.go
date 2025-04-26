package plugin

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/rogpeppe/go-internal/testenv"
	"github.com/stretchr/testify/require"
)

func TestOpen(t *testing.T) {
	tmpDir := t.TempDir()

	code := `
		package main

		import (
			"context"
		)

		type Config struct {}

		type Plugin struct {}

		func New(_ Config) *Plugin {
			return &Plugin{}, nil
		}

		func (p *Plugin) Load(ctx context.Context) error {
			return nil
		}

		func (p *Plugin) Unload(ctx context.Context) error {
			return nil
		}
	`

	src := filepath.Join(tmpDir, "plugin.go")
	dist := filepath.Join(tmpDir, "plugin.so")

	err := os.WriteFile(src, []byte(code), 0644)
	require.NoError(t, err)

	cmd := exec.Command(testenv.GoToolPath(t), "build", "-buildmode=plugin", "-o", dist, src)
	cmd.Env = os.Environ()
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	p, err := Open(dist, map[string]any{"name": "test-plugin"})
	require.NoError(t, err)
	require.NotNil(t, p)
}
