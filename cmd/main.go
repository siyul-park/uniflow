package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/iancoleman/strcase"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/siyul-park/uniflow/cli"
	"github.com/siyul-park/uniflow/hook"
	"github.com/siyul-park/uniflow/scheme"
	"github.com/siyul-park/uniflow/spec"
	"github.com/siyul-park/uniflow/store"
	"github.com/siyul-park/uniflow/testing"
	"github.com/siyul-park/uniflow/value"
	"github.com/spf13/afero"
)

const configFile = ".uniflow.toml"

var k = koanf.New(".")

func init() {
	if err := k.Set(cli.EnvCollectionSpecs, "specs"); err != nil {
		log.Fatal(err)
	}
	if err := k.Set(cli.EnvCollectionValues, "values"); err != nil {
		log.Fatal(err)
	}

	_ = k.Load(file.Provider(configFile), toml.Parser())

	if err := k.Load(env.Provider("", ".", func(s string) string {
		return strcase.ToDelimited(s, '.')
	}), nil); err != nil {
		log.Fatal(err)
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	//databaseURL := k.String(cli.EnvDatabaseURL)
	//databaseName := k.String(cli.EnvDatabaseName)
	collectionNodes := k.String(cli.EnvCollectionSpecs)
	collectionValues := k.String(cli.EnvCollectionValues)

	source := store.NewSource()
	defer source.Close()

	specStore, err := source.Open(collectionNodes)
	if err != nil {
		log.Fatal(err)
	}
	valueStore, err := source.Open(collectionValues)
	if err != nil {
		log.Fatal(err)
	}

	if err := specStore.Index(ctx, []string{spec.KeyNamespace, spec.KeyName}, store.IndexOptions{
		Unique: true,
		Filter: map[string]any{spec.KeyName: map[string]any{"$exists": true}},
	}); err != nil {
		log.Fatal(err)
	}
	if err := valueStore.Index(ctx, []string{value.KeyNamespace, value.KeyName}, store.IndexOptions{
		Unique: true,
		Filter: map[string]any{value.KeyName: map[string]any{"$exists": true}},
	}); err != nil {
		log.Fatal(err)
	}

	runner := testing.NewRunner()

	schemeBuilder := scheme.NewBuilder()
	hookBuilder := hook.NewBuilder()

	scheme, err := schemeBuilder.Build()
	if err != nil {
		log.Fatal(err)
	}
	hook, err := hookBuilder.Build()
	if err != nil {
		log.Fatal(err)
	}

	fs := afero.NewOsFs()

	cmd := cli.NewCommand(cli.Config{
		Use:   "uniflow",
		Short: "A high-performance, extremely flexible, and easily extensible universal workflow engine.",
		FS:    fs,
	})
	cmd.AddCommand(cli.NewStartCommand(cli.StartConfig{
		Scheme:     scheme,
		Hook:       hook,
		SpecStore:  specStore,
		ValueStore: valueStore,
		FS:         fs,
	}))
	cmd.AddCommand(cli.NewTestCommand(cli.TestConfig{
		Runner:     runner,
		Scheme:     scheme,
		Hook:       hook,
		SpecStore:  specStore,
		ValueStore: valueStore,
		FS:         fs,
	}))
	cmd.AddCommand(cli.NewApplyCommand(cli.ApplyConfig{
		SpecStore:  specStore,
		ValueStore: valueStore,
		FS:         fs,
	}))
	cmd.AddCommand(cli.NewDeleteCommand(cli.DeleteConfig{
		SpecStore:  specStore,
		ValueStore: valueStore,
		FS:         fs,
	}))
	cmd.AddCommand(cli.NewGetCommand(cli.GetConfig{
		SpecStore:  specStore,
		ValueStore: valueStore,
	}))

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
