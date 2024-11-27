package main

import (
	"context"
	"github.com/siyul-park/uniflow/cmd/pkg/driver"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/siyul-park/uniflow/cmd/pkg/cli"
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
	"github.com/spf13/viper"
)

const configFile = ".uniflow.toml"

const (
	opCreateCharts = "charts.create"
	opReadCharts   = "charts.read"
	opUpdateCharts = "charts.update"
	opDeleteCharts = "charts.delete"

	opCreateSpecs = "specs.create"
	opReadSpecs   = "specs.read"
	opUpdateSpecs = "specs.update"
	opDeleteSpecs = "specs.delete"

	opCreateSecrets = "secrets.create"
	opReadSecrets   = "secrets.read"
	opUpdateSecrets = "secrets.update"
	opDeleteSecrets = "secrets.delete"
)

func init() {
	viper.SetDefault(cli.EnvCollectionSpecs, "specs")
	viper.SetDefault(cli.EnvCollectionSecrets, "secrets")
	viper.SetDefault(cli.EnvCollectionCharts, "charts")

	viper.SetConfigFile(configFile)
	viper.AutomaticEnv()
	viper.ReadInConfig()
}

func main() {
	ctx := context.Background()

	databaseURL := viper.GetString(cli.EnvDatabaseURL)
	databaseName := viper.GetString(cli.EnvDatabaseName)
	collectionNodes := viper.GetString(cli.EnvCollectionSpecs)
	collectionSecrets := viper.GetString(cli.EnvCollectionSecrets)
	collectionCharts := viper.GetString(cli.EnvCollectionCharts)

	d := driver.NewInMemoryDriver()
	defer d.Close(ctx)

	if strings.HasPrefix(databaseURL, "memongodb://") || strings.HasPrefix(databaseURL, "mongodb://") {
		var err error
		if d, err = driver.NewMongoDriver(ctx, databaseURL, databaseName); err != nil {
			log.Fatal(err)
		}
	}

	specStore, err := d.SpecStore(ctx, collectionNodes)
	if err != nil {
		log.Fatal(err)
	}
	secretStore, err := d.SecretStore(ctx, collectionSecrets)
	if err != nil {
		log.Fatal(err)
	}
	chartStore, err := d.ChartStore(ctx, collectionCharts)
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

	operators := map[string]any{
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

	schemeBuilder.Register(control.AddToScheme(languages, cel.Language))
	schemeBuilder.Register(io.AddToScheme(io.NewOSFileSystem()))
	schemeBuilder.Register(network.AddToScheme())
	schemeBuilder.Register(system.AddToScheme(operators))

	hookBuilder.Register(control.AddToHook())
	hookBuilder.Register(network.AddToHook())

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
