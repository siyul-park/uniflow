package main

import (
	"context"

	"github.com/siyul-park/uniflow/pkg/plugin"
)

type Config struct {
	Foo string `json:"foo" validate:"required"`
}

type Plugin struct {
}

var _ plugin.Plugin = (*Plugin)(nil)

func New(config Config) plugin.Plugin {
	return &Plugin{}
}

func (p *Plugin) Load(ctx context.Context) error {
	return nil
}

func (p *Plugin) Unload(ctx context.Context) error {
	return nil
}
