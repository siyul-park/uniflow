package main

import (
	"context"
	"net/url"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/iancoleman/strcase"
	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/parsers/hjson"
	"github.com/knadh/koanf/parsers/toml/v2"
	koanfyaml "github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/pkg/errors"
	"github.com/spf13/afero"

	"github.com/siyul-park/uniflow/internal/cli"
	"github.com/siyul-park/uniflow/pkg/driver"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/language"
	"github.com/siyul-park/uniflow/pkg/language/json"
	"github.com/siyul-park/uniflow/pkg/language/text"
	"github.com/siyul-park/uniflow/pkg/language/yaml"
	"github.com/siyul-park/uniflow/pkg/meta"
	"github.com/siyul-park/uniflow/pkg/plugin"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/testing"
	"github.com/siyul-park/uniflow/pkg/value"
)

const (
	prefix = "UNIFLOW_"

	keyConfig           = "config"
	keyRuntimeLanguage  = "runtime.language"
	KeyRuntimeNamespace = "runtime.namespace"
	keyEnvironment      = "environment"
	keyDatabaseURL      = "database.url"
	keyCollectionSpecs  = "collection.specs"
	keyCollectionValues = "collection.values"
	keyPlugins          = "plugins"
)

var k = koanf.New(".")

func init() {
	cli.Fatal(k.Set(keyConfig, ".uniflow.toml"))
	cli.Fatal(k.Set(KeyRuntimeNamespace, meta.DefaultNamespace))
	cli.Fatal(k.Set(keyDatabaseURL, "memory://"))
	cli.Fatal(k.Set(keyCollectionSpecs, "specs"))
	cli.Fatal(k.Set(keyCollectionValues, "values"))

	cli.Fatal(k.Load(env.Provider(prefix, ".", func(s string) string {
		return strcase.ToDelimited(strings.TrimPrefix(s, prefix), '.')
	}), nil))

	config := k.String(keyConfig)

	var parser koanf.Parser
	switch strings.ToLower(filepath.Ext(config)) {
	case ".toml":
		parser = toml.Parser()
	case ".yaml", ".yml":
		parser = koanfyaml.Parser()
	case ".json", ".hjson":
		parser = hjson.Parser()
	case ".env":
		parser = dotenv.ParserEnv(prefix, ".", func(s string) string {
			return strcase.ToDelimited(strings.TrimPrefix(s, prefix), '.')
		})
	default:
		cli.Fatal(errors.New("invalid config file extension"))
	}

	_ = k.Load(file.Provider(config), parser)
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

	languageRegistry.SetDefault(k.String(keyRuntimeLanguage))

	cli.Fatal(languageRegistry.Register(text.Language, text.NewCompiler()))
	cli.Fatal(languageRegistry.Register(json.Language, json.NewCompiler()))
	cli.Fatal(languageRegistry.Register(yaml.Language, yaml.NewCompiler()))

	pluginRegistry := plugin.NewRegistry()
	defer pluginRegistry.Unload(ctx)

	driverProxy := driver.NewProxy(nil)
	defer driverProxy.Close()

	agent := runtime.NewAgent()
	defer agent.Close()

	for _, cfg := range k.Slices(keyPlugins) {
		p := cli.Must(plugin.Open(cfg.String("path"), cfg.Get("config")))
		cli.Fatal(pluginRegistry.Register(p))
	}
	cli.Fatal(pluginRegistry.Inject(testingRunner, schemeBuilder, hookBuilder, pluginRegistry, driverRegistry, languageRegistry, driverProxy, agent))
	cli.Fatal(pluginRegistry.Load(ctx))

	sc := cli.Must(schemeBuilder.Build())
	hk := cli.Must(hookBuilder.Build())

	dsn := cli.Must(url.Parse(k.String(keyDatabaseURL)))

	drv := cli.Must(driverRegistry.Lookup(dsn.Scheme))
	defer drv.Close()

	driverProxy.Wrap(drv)

	conn := cli.Must(drv.Open(dsn.String()))
	defer conn.Close()

	specStore := cli.Must(conn.Load(k.String(keyCollectionSpecs)))
	valueStore := cli.Must(conn.Load(k.String(keyCollectionValues)))

	cli.Fatal(specStore.Index(ctx, []string{spec.KeyNamespace, spec.KeyName}, driver.IndexOptions{
		Unique: true,
		Filter: map[string]any{spec.KeyName: map[string]any{"$exists": true}},
	}))
	cli.Fatal(valueStore.Index(ctx, []string{value.KeyNamespace, value.KeyName}, driver.IndexOptions{
		Unique: true,
		Filter: map[string]any{value.KeyName: map[string]any{"$exists": true}},
	}))

	namespace := k.String(KeyRuntimeNamespace)
	environment := k.StringMap(keyEnvironment)

	fs := afero.NewOsFs()

	cmd := cli.NewCommand(cli.Config{
		Use:   "uniflow",
		Short: "A high-performance, extremely flexible, and easily extensible universal workflow engine.",
		FS:    fs,
	})
	cmd.AddCommand(cli.NewStartCommand(cli.StartConfig{
		Namespace:   namespace,
		Environment: environment,
		Agent:       agent,
		Scheme:      sc,
		Hook:        hk,
		SpecStore:   specStore,
		ValueStore:  valueStore,
		FS:          fs,
	}))
	cmd.AddCommand(cli.NewTestCommand(cli.TestConfig{
		Namespace:   namespace,
		Environment: environment,
		Runner:      testingRunner,
		Scheme:      sc,
		Hook:        hk,
		SpecStore:   specStore,
		ValueStore:  valueStore,
		FS:          fs,
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
