package main

import (
	"context"
	"fmt"
	"os"

	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/spf13/viper"
)

const (
	configFile = ".uniflow.toml"
)

func init() {
	viper.SetConfigFile(configFile)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func main() {
	if err := execute(); err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
}

func execute() error {
	ctx := context.Background()

	sb := scheme.NewBuilder()
	hb := hook.NewBuilder()

	sc, err := sb.Build()
	if err != nil {
		return err
	}
	hk, err := hb.Build()
	if err != nil {
		return err
	}

	db, err := connectDatabase(ctx)
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
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}
