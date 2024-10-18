package main

import (
	"context"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/siyul-park/uniflow/cmd/pkg/cli"
	mongochart "github.com/siyul-park/uniflow/driver/mongo/pkg/chart"
	mongosecret "github.com/siyul-park/uniflow/driver/mongo/pkg/secret"
	mongoserver "github.com/siyul-park/uniflow/driver/mongo/pkg/server"
	mongospec "github.com/siyul-park/uniflow/driver/mongo/pkg/spec"
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
	"github.com/siyul-park/uniflow/pkg/chart"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const configFile = ".uniflow.toml"

const (
	flagDatabaseURL       = "database.url"
	flagDatabaseName      = "database.name"
	flagCollectionCharts  = "collection.charts"
	flagCollectionNodes   = "collection.nodes"
	flagCollectionSecrets = "collection.secrets"
)

const (
	opCreateCharts = "charts.create"
	opReadCharts   = "charts.read"
	opUpdateCharts = "charts.update"
	opDeleteCharts = "charts.delete"

	opCreateNodes = "nodes.create"
	opReadNodes   = "nodes.read"
	opUpdateNodes = "nodes.update"
	opDeleteNodes = "nodes.delete"

	opCreateSecrets = "secrets.create"
	opReadSecrets   = "secrets.read"
	opUpdateSecrets = "secrets.update"
	opDeleteSecrets = "secrets.delete"
)

func init() {
	viper.SetDefault(flagCollectionCharts, "charts")
	viper.SetDefault(flagCollectionNodes, "nodes")
	viper.SetDefault(flagCollectionSecrets, "secrets")

	viper.SetConfigFile(configFile)
	viper.AutomaticEnv()
	viper.ReadInConfig()
}

func main() {
	ctx := context.Background()

	databaseURL := viper.GetString(flagDatabaseURL)
	databaseName := viper.GetString(flagDatabaseName)
	collectionCharts := viper.GetString(flagCollectionCharts)
	collectionNodes := viper.GetString(flagCollectionNodes)
	collectionSecrets := viper.GetString(flagCollectionSecrets)

	if strings.HasPrefix(databaseURL, "memongodb://") {
		server := mongoserver.New()
		defer server.Stop()

		databaseURL = server.URI()
	}

	var chartStore chart.Store
	var specStore spec.Store
	var secretStore secret.Store
	if strings.HasPrefix(databaseURL, "mongodb://") {
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		opts := options.Client().ApplyURI(databaseURL).SetServerAPIOptions(serverAPI)

		client, err := mongo.Connect(ctx, opts)
		if err != nil {
			log.Fatal(err)
		}
		defer client.Disconnect(ctx)

		collection := client.Database(databaseName).Collection(collectionCharts)
		chartStore = mongochart.NewStore(collection)
		if err := chartStore.(*mongochart.Store).Index(ctx); err != nil {
			log.Fatal(err)
		}

		collection = client.Database(databaseName).Collection(collectionNodes)
		specStore = mongospec.NewStore(collection)
		if err := specStore.(*mongospec.Store).Index(ctx); err != nil {
			log.Fatal(err)
		}

		collection = client.Database(databaseName).Collection(collectionSecrets)
		secretStore = mongosecret.NewStore(collection)
		if err := secretStore.(*mongosecret.Store).Index(ctx); err != nil {
			log.Fatal(err)
		}
	} else {
		specStore = spec.NewStore()
		secretStore = secret.NewStore()
	}

	schemeBuilder := scheme.NewBuilder()
	hookBuilder := hook.NewBuilder()

	langs := language.NewModule()
	langs.Store(text.Language, text.NewCompiler())
	langs.Store(json.Language, json.NewCompiler())
	langs.Store(yaml.Language, yaml.NewCompiler())
	langs.Store(cel.Language, cel.NewCompiler())
	langs.Store(javascript.Language, javascript.NewCompiler())
	langs.Store(typescript.Language, typescript.NewCompiler())

	nativeTable := system.NewNativeTable()
	nativeTable.Store(opCreateCharts, system.CreateResource(chartStore))
	nativeTable.Store(opReadCharts, system.ReadResource(chartStore))
	nativeTable.Store(opUpdateCharts, system.UpdateResource(chartStore))
	nativeTable.Store(opDeleteCharts, system.DeleteResource(chartStore))
	nativeTable.Store(opCreateNodes, system.CreateResource(specStore))
	nativeTable.Store(opReadNodes, system.ReadResource(specStore))
	nativeTable.Store(opUpdateNodes, system.UpdateResource(specStore))
	nativeTable.Store(opDeleteNodes, system.DeleteResource(specStore))
	nativeTable.Store(opCreateSecrets, system.CreateResource(secretStore))
	nativeTable.Store(opReadSecrets, system.ReadResource(secretStore))
	nativeTable.Store(opUpdateSecrets, system.UpdateResource(secretStore))
	nativeTable.Store(opDeleteSecrets, system.DeleteResource(secretStore))

	schemeBuilder.Register(control.AddToScheme(langs, cel.Language))
	schemeBuilder.Register(io.AddToScheme(io.NewOSFileSystem()))
	schemeBuilder.Register(network.AddToScheme())
	schemeBuilder.Register(system.AddToScheme(nativeTable))

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
