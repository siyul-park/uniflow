package main

import (
	"context"
	"os"

	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/plugin/pkg/control"
	"github.com/siyul-park/uniflow/plugin/pkg/network"
)

func execute(ctx context.Context, databaseURL, databaseName string) error {
	sb := scheme.NewBuilder()
	hb := hook.NewBuilder()

	sb.Register(control.AddToScheme())
	sb.Register(network.AddToScheme())

	hb.Register(network.AddToHook())

	sc, err := sb.Build()
	if err != nil {
		return err
	}
	hk, err := hb.Build()
	if err != nil {
		return err
	}

	db, err := connectDatabase(ctx, databaseURL, databaseName)
	if err != nil {
		return err
	}

	curDir, err := os.Getwd()
	if err != nil {
		return err
	}
	fsys := os.DirFS(curDir)

	cmd := NewCommand(Config{
		Scheme:   sc,
		Hook:     hk,
		Database: db,
		FS:       fsys,
	})

	return cmd.Execute()
}
