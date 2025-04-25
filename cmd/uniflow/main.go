package main

import (
	"context"
	"log"
	"net/url"
	"os/signal"
	"syscall"

	"github.com/iancoleman/strcase"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/afero"

	"github.com/siyul-park/uniflow/internal/cli"
	"github.com/siyul-park/uniflow/pkg/driver"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/testing"
	"github.com/siyul-park/uniflow/pkg/value"
)

const (
	envDatabaseURL      = "database.url"
	envCollectionSpecs  = "collection.specs"
	envCollectionValues = "collection.values"
)

const configFile = ".uniflow.toml"

var config = koanf.New(".")

func init() {
	must(config.Set(envCollectionSpecs, "specs"))
	must(config.Set(envCollectionValues, "values"))

	_ = config.Load(file.Provider(configFile), toml.Parser())

	must(config.Load(env.Provider("", ".", func(s string) string {
		return strcase.ToDelimited(s, '.')
	}), nil))
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	databaseURL := config.String(envDatabaseURL)
	specsCollection := config.String(envCollectionSpecs)
	valuesCollection := config.String(envCollectionValues)

	dsn, err := url.Parse(databaseURL)
	must(err)

	driverRegistry := driver.NewRegistry()
	defer driverRegistry.Close()

	must(driverRegistry.Register("mem", driver.New()))

	fs := afero.NewOsFs()
	runner := testing.NewRunner()

	schemeBuilder := scheme.NewBuilder()
	hookBuilder := hook.NewBuilder()

	drv, err := driverRegistry.Lookup(dsn.Scheme)
	must(err)

	conn, err := drv.Open(databaseURL)
	must(err)

	specStore, err := conn.Load(specsCollection)
	must(err)

	valueStore, err := conn.Load(valuesCollection)
	must(err)

	scheme, err := schemeBuilder.Build()
	must(err)

	hook, err := hookBuilder.Build()
	must(err)

	must(specStore.Index(ctx, []string{spec.KeyNamespace, spec.KeyName}, driver.IndexOptions{
		Unique: true,
		Filter: map[string]any{spec.KeyName: map[string]any{"$exists": true}},
	}))
	must(valueStore.Index(ctx, []string{value.KeyNamespace, value.KeyName}, driver.IndexOptions{
		Unique: true,
		Filter: map[string]any{value.KeyName: map[string]any{"$exists": true}},
	}))

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

	must(cmd.Execute())
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
