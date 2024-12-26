package main

import (
	"context"
	"log"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/siyul-park/uniflow/cmd/pkg/cli"
	"github.com/siyul-park/uniflow/cmd/pkg/driver"
	"github.com/siyul-park/uniflow/ext/pkg/control"
	"github.com/siyul-park/uniflow/ext/pkg/io"
	"github.com/siyul-park/uniflow/ext/pkg/language"
	"github.com/siyul-park/uniflow/ext/pkg/language/cel"
	"github.com/siyul-park/uniflow/ext/pkg/language/javascript"
	"github.com/siyul-park/uniflow/ext/pkg/language/json"
	"github.com/siyul-park/uniflow/ext/pkg/language/text"
	"github.com/siyul-park/uniflow/ext/pkg/language/typescript"
	"github.com/siyul-park/uniflow/ext/pkg/language/yaml"
	"github.com/siyul-park/uniflow/ext/pkg/network"
	"github.com/siyul-park/uniflow/ext/pkg/system"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/spf13/afero"
)

const configFile = ".uniflow.toml"

const (
	topicSpecs   = "specs"
	topicSecrets = "secrets"
	topicCharts  = "charts"

	opCreateSpecs = "specs.create"
	opReadSpecs   = "specs.read"
	opUpdateSpecs = "specs.update"
	opDeleteSpecs = "specs.delete"

	opCreateSecrets = "secrets.create"
	opReadSecrets   = "secrets.read"
	opUpdateSecrets = "secrets.update"
	opDeleteSecrets = "secrets.delete"

	opCreateCharts = "charts.create"
	opReadCharts   = "charts.read"
	opUpdateCharts = "charts.update"
	opDeleteCharts = "charts.delete"
)

var k = koanf.New(".")

func init() {
	if err := k.Set(cli.EnvCollectionSpecs, "specs"); err != nil {
		log.Fatal(err)
	}
	if err := k.Set(cli.EnvCollectionSecrets, "secrets"); err != nil {
		log.Fatal(err)
	}
	if err := k.Set(cli.EnvCollectionCharts, "charts"); err != nil {
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
	ctx := context.Background()

	databaseURL := k.String(cli.EnvDatabaseURL)
	databaseName := k.String(cli.EnvDatabaseName)
	collectionNodes := k.String(cli.EnvCollectionSpecs)
	collectionSecrets := k.String(cli.EnvCollectionSecrets)
	collectionCharts := k.String(cli.EnvCollectionCharts)

	drv := driver.NewInMemoryDriver()
	defer drv.Close(ctx)

	if strings.HasPrefix(databaseURL, "memongodb://") || strings.HasPrefix(databaseURL, "mongodb://") {
		var err error
		if drv, err = driver.NewMongoDriver(ctx, databaseURL, databaseName); err != nil {
			log.Fatal(err)
		}
	}

	specStore, err := drv.SpecStore(ctx, collectionNodes)
	if err != nil {
		log.Fatal(err)
	}
	secretStore, err := drv.SecretStore(ctx, collectionSecrets)
	if err != nil {
		log.Fatal(err)
	}
	chartStore, err := drv.ChartStore(ctx, collectionCharts)
	if err != nil {
		log.Fatal(err)
	}

	schemeBuilder := scheme.NewBuilder()
	hookBuilder := hook.NewBuilder()

	languages := language.NewModule()
	languages.Store(text.Language, text.NewCompiler())
	languages.Store(json.Language, json.NewCompiler())
	languages.Store(yaml.Language, yaml.NewCompiler())
	languages.Store(cel.Language, cel.NewCompiler())
	languages.Store(javascript.Language, javascript.NewCompiler())
	languages.Store(typescript.Language, typescript.NewCompiler())

	signals := map[string]any{
		topicSpecs:   system.WatchResource(specStore),
		topicSecrets: system.WatchResource(secretStore),
		topicCharts:  system.WatchResource(chartStore),
	}
	calls := map[string]any{
		opCreateSpecs:   system.CreateResource(specStore),
		opReadSpecs:     system.ReadResource(specStore),
		opUpdateSpecs:   system.UpdateResource(specStore),
		opDeleteSpecs:   system.DeleteResource(specStore),
		opCreateSecrets: system.CreateResource(secretStore),
		opReadSecrets:   system.ReadResource(secretStore),
		opUpdateSecrets: system.UpdateResource(secretStore),
		opDeleteSecrets: system.DeleteResource(secretStore),
		opCreateCharts:  system.CreateResource(chartStore),
		opReadCharts:    system.ReadResource(chartStore),
		opUpdateCharts:  system.UpdateResource(chartStore),
		opDeleteCharts:  system.DeleteResource(chartStore),
	}

	systemAddToScheme := system.AddToScheme()

	for topic, signal := range signals {
		if err := systemAddToScheme.SetSignal(topic, signal); err != nil {
			log.Fatal(err)
		}
	}
	for opcode, call := range calls {
		if err := systemAddToScheme.SetCall(opcode, call); err != nil {
			log.Fatal(err)
		}
	}

	schemeBuilder.Register(control.AddToScheme(languages, cel.Language))
	schemeBuilder.Register(io.AddToScheme(io.NewOSFileSystem()))
	schemeBuilder.Register(network.AddToScheme())
	schemeBuilder.Register(systemAddToScheme)

	hookBuilder.Register(network.AddToHook())
	hookBuilder.Register(system.AddToHook())

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
		Scheme:      scheme,
		Hook:        hook,
		ChartStore:  chartStore,
		SpecStore:   specStore,
		SecretStore: secretStore,
		FS:          fs,
	}))

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
