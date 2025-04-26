package main

import (
	"context"
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
	"github.com/siyul-park/uniflow/pkg/language"
	"github.com/siyul-park/uniflow/pkg/language/json"
	"github.com/siyul-park/uniflow/pkg/language/text"
	"github.com/siyul-park/uniflow/pkg/language/yaml"
	"github.com/siyul-park/uniflow/pkg/plugin"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/testing"
	"github.com/siyul-park/uniflow/pkg/value"
)

const (
	envDatabaseURL      = "database.url"
	envCollectionSpecs  = "collection.specs"
	envCollectionValues = "collection.values"
	envLanguageDefault  = "language.default"
	envPlugin           = "plugin"
)

const configFile = ".uniflow.toml"

var k = koanf.New(".")

func init() {
	cli.Fatal(k.Set(envDatabaseURL, "memory://"))
	cli.Fatal(k.Set(envCollectionSpecs, "specs"))
	cli.Fatal(k.Set(envCollectionValues, "values"))

	_ = k.Load(file.Provider(configFile), toml.Parser())

	cli.Fatal(k.Load(env.Provider("", ".", func(s string) string {
		return strcase.ToDelimited(s, '.')
	}), nil))
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	testingRunner := testing.NewRunner()

	schemeBuilder := scheme.NewBuilder()
	hookBuilder := hook.NewBuilder()

	driverRegistry := driver.NewRegistry()
	defer driverRegistry.Close()

	cli.Fatal(driverRegistry.Register("memory", driver.New()))

	languageRegistry := language.NewRegistry()
	defer languageRegistry.Close()

	languageRegistry.SetDefault(k.String(envLanguageDefault))

	cli.Fatal(languageRegistry.Register(text.Language, text.NewCompiler()))
	cli.Fatal(languageRegistry.Register(json.Language, json.NewCompiler()))
	cli.Fatal(languageRegistry.Register(yaml.Language, yaml.NewCompiler()))

	pluginRegistry := plugin.NewRegistry()
	defer pluginRegistry.Unload(ctx)

	for _, cfg := range k.Slices(envPlugin) {
		p := cli.Must(plugin.Open(cfg.String("path"), cfg.Get("config")))
		cli.Fatal(pluginRegistry.Register(p))
	}
	cli.Fatal(pluginRegistry.Inject(testingRunner, schemeBuilder, hookBuilder, driverRegistry, languageRegistry))
	cli.Fatal(pluginRegistry.Load(ctx))

	sc := cli.Must(schemeBuilder.Build())
	hk := cli.Must(hookBuilder.Build())

	dsn := cli.Must(url.Parse(k.String(envDatabaseURL)))

	drv := cli.Must(driverRegistry.Lookup(dsn.Scheme))
	defer drv.Close()

	conn := cli.Must(drv.Open(dsn.String()))
	defer conn.Close()

	specStore := cli.Must(conn.Load(k.String(envCollectionSpecs)))
	valueStore := cli.Must(conn.Load(k.String(envCollectionValues)))

	cli.Fatal(specStore.Index(ctx, []string{spec.KeyNamespace, spec.KeyName}, driver.IndexOptions{
		Unique: true,
		Filter: map[string]any{spec.KeyName: map[string]any{"$exists": true}},
	}))
	cli.Fatal(valueStore.Index(ctx, []string{value.KeyNamespace, value.KeyName}, driver.IndexOptions{
		Unique: true,
		Filter: map[string]any{value.KeyName: map[string]any{"$exists": true}},
	}))

	fs := afero.NewOsFs()

	cmd := cli.NewCommand(cli.Config{
		Use:   "uniflow",
		Short: "A high-performance, extremely flexible, and easily extensible universal workflow engine.",
		FS:    fs,
	})
	cmd.AddCommand(cli.NewStartCommand(cli.StartConfig{
		Scheme:     sc,
		Hook:       hk,
		SpecStore:  specStore,
		ValueStore: valueStore,
		FS:         fs,
	}))
	cmd.AddCommand(cli.NewTestCommand(cli.TestConfig{
		Runner:     testingRunner,
		Scheme:     sc,
		Hook:       hk,
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

	cli.Fatal(cmd.Execute())
}
