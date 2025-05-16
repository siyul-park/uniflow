package plugin

import (
	"context"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLoader_Load(t *testing.T) {
	fs := afero.NewMemMapFs()

	file, err := fs.Create("main.go")
	require.NoError(t, err)
	defer file.Close()

	_, err = file.WriteString(`
package main

import "context"

type Plugin struct {
}

func New() *Plugin {
	return &Plugin{}
}

func (p *Plugin) Name() string {
	return ""
}

func (p *Plugin) Version() string {
	return ""
}

func (p *Plugin) Load(ctx context.Context) error {
	return nil
}

func (p *Plugin) Unload() error {
	return nil
}

`)
	require.NoError(t, err)

	ld := NewLoader(fs)

	p, err := ld.Open("main.go", nil)
	require.NoError(t, err)

	err = p.Load(context.Background())
	require.NoError(t, err)
}
