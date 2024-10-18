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
	"github.com/siyul-park/uniflow/pkg/chart"
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

	fs := afero.NewOsFs()

	cmd := cli.NewCommand(cli.Config{
		Use:   "uniflowctl",
		Short: "A high-performance, extremely flexible, and easily extensible universal workflow engine.",
	})
	cmd.AddCommand(cli.NewApplyCommand(cli.ApplyConfig{
		SpecStore:   specStore,
		SecretStore: secretStore,
		FS:          fs,
	}))
	cmd.AddCommand(cli.NewDeleteCommand(cli.DeleteConfig{
		SpecStore:   specStore,
		SecretStore: secretStore,
		FS:          fs,
	}))
	cmd.AddCommand(cli.NewGetCommand(cli.GetConfig{
		SpecStore:   specStore,
		SecretStore: secretStore,
	}))

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
