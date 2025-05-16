package plugin

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestLoader_Load(t *testing.T) {
	fs := afero.NewMemMapFs()

	file, err := fs.Create("main.go")
	require.NoError(t, err)
	defer file.Close()

	_, err = file.WriteString(`
package main

import (
	"context"

	"github.com/siyul-park/uniflow/pkg/plugin"
)

type Plugin struct {
}

var _ plugin.Plugin = (*Plugin)(nil)

func New() *Plugin {
	return &Plugin{}
}

func (p *Plugin) Name() string {
	return "test"
}

func (p *Plugin) Version() string {
	return "test"
}

func (p *Plugin) Load(_ context.Context) error {
	return nil
}

func (p *Plugin) Unload(_ context.Context) error {
	return nil
}

`)
	require.NoError(t, err)

	ld := NewLoader(fs)

	p, err := ld.Open("main.go", nil)
	require.NoError(t, err)

	name := p.Name()
	require.Equal(t, "test", name)

	version := p.Version()
	require.Equal(t, "test", version)

	err = p.Load(context.Background())
	require.NoError(t, err)

	err = p.Unload(context.Background())
	require.NoError(t, err)
}
