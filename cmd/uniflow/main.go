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

	"github.com/siyul-park/uniflow/internal/cmd"
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
	cmd.Fatal(k.Set(keyConfig, ".uniflow.toml"))
	cmd.Fatal(k.Set(KeyRuntimeNamespace, meta.DefaultNamespace))
	cmd.Fatal(k.Set(keyDatabaseURL, "memory://"))
	cmd.Fatal(k.Set(keyCollectionSpecs, "specs"))
	cmd.Fatal(k.Set(keyCollectionValues, "values"))

	cmd.Fatal(k.Load(env.Provider(prefix, ".", func(s string) string {
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
		cmd.Fatal(errors.New("invalid config file extension"))
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

	cmd.Fatal(driverRegistry.Register("memory", driver.New()))

	languageRegistry := language.NewRegistry()
	defer languageRegistry.Close()

	languageRegistry.SetDefault(k.String(keyRuntimeLanguage))

	cmd.Fatal(languageRegistry.Register(text.Language, text.NewCompiler()))
	cmd.Fatal(languageRegistry.Register(json.Language, json.NewCompiler()))
	cmd.Fatal(languageRegistry.Register(yaml.Language, yaml.NewCompiler()))

	pluginRegistry := plugin.NewRegistry()
	defer pluginRegistry.Unload(ctx)

	connProxy := driver.NewConnProxy(nil)
	defer connProxy.Close()

	agent := runtime.NewAgent()
	defer agent.Close()

	fs := afero.NewOsFs()

	pluginLoader := plugin.NewLoader(fs)

	for _, cfg := range k.Slices(keyPlugins) {
		e := map[string]string{}
		for _, key := range cfg.Keys() {
			e[strcase.ToScreamingSnake(key)] = cfg.String(key)
		}
		p := cmd.Must(pluginLoader.Open(cfg.String("path"), plugin.LoadOptions{
			Environment: e,
			Arguments:   []any{cfg.Get("config")},
		}))
		cmd.Fatal(pluginRegistry.Register(p))
	}
	for _, dep := range []any{testingRunner, connProxy, agent, fs, schemeBuilder, hookBuilder, pluginRegistry, driverRegistry, languageRegistry} {
		cmd.Must(pluginRegistry.Inject(dep))
	}
	cmd.Fatal(pluginRegistry.Load(ctx))

	sc := cmd.Must(schemeBuilder.Build())
	hk := cmd.Must(hookBuilder.Build())

	dsn := cmd.Must(url.Parse(k.String(keyDatabaseURL)))

	drv := cmd.Must(driverRegistry.Lookup(dsn.Scheme))
	defer drv.Close()

	conn := cmd.Must(drv.Open(dsn.String()))
	defer conn.Close()

	connAlias := driver.NewConnAlias(conn)
	defer connAlias.Close()

	connAlias.Alias(k.String(keyCollectionSpecs), "specs")
	connAlias.Alias(k.String(keyCollectionValues), "values")

	connProxy.Wrap(connAlias)

	specStore := cmd.Must(conn.Load(k.String(keyCollectionSpecs)))
	valueStore := cmd.Must(conn.Load(k.String(keyCollectionValues)))

	cmd.Fatal(specStore.Index(ctx, []string{spec.KeyNamespace, spec.KeyName}, driver.IndexOptions{
		Unique: true,
		Filter: map[string]any{spec.KeyName: map[string]any{"$exists": true}},
	}))
	cmd.Fatal(valueStore.Index(ctx, []string{value.KeyNamespace, value.KeyName}, driver.IndexOptions{
		Unique: true,
		Filter: map[string]any{value.KeyName: map[string]any{"$exists": true}},
	}))

	namespace := k.String(KeyRuntimeNamespace)
	environment := k.StringMap(keyEnvironment)

	root := cmd.NewCommand(cmd.Config{
		Use:   "uniflow",
		Short: "A high-performance, extremely flexible, and easily extensible universal workflow engine.",
		FS:    fs,
	})
	root.AddCommand(cmd.NewStartCommand(cmd.StartConfig{
		Namespace:   namespace,
		Environment: environment,
		Agent:       agent,
		Scheme:      sc,
		Hook:        hk,
		SpecStore:   specStore,
		ValueStore:  valueStore,
		FS:          fs,
	}))
	root.AddCommand(cmd.NewTestCommand(cmd.TestConfig{
		Namespace:   namespace,
		Environment: environment,
		Runner:      testingRunner,
		Scheme:      sc,
		Hook:        hk,
		SpecStore:   specStore,
		ValueStore:  valueStore,
		FS:          fs,
	}))
	root.AddCommand(cmd.NewApplyCommand(cmd.ApplyConfig{
		SpecStore:  specStore,
		ValueStore: valueStore,
		FS:         fs,
	}))
	root.AddCommand(cmd.NewDeleteCommand(cmd.DeleteConfig{
		SpecStore:  specStore,
		ValueStore: valueStore,
		FS:         fs,
	}))
	root.AddCommand(cmd.NewGetCommand(cmd.GetConfig{
		SpecStore:  specStore,
		ValueStore: valueStore,
	}))

	cmd.Fatal(root.Execute())
}
