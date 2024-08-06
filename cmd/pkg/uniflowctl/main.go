package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/siyul-park/uniflow/cmd/pkg/cli"
	mongosecret "github.com/siyul-park/uniflow/driver/mongo/pkg/secret"
	mongoserver "github.com/siyul-park/uniflow/driver/mongo/pkg/server"
	mongospec "github.com/siyul-park/uniflow/driver/mongo/pkg/spec"
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
	flagCollectionNodes   = "collection.nodes"
	flagCollectionSecrets = "collection.secrets"
)

func init() {
	viper.SetConfigFile(configFile)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func main() {
	ctx := context.Background()

	databaseURL := viper.GetString(flagDatabaseURL)
	databaseName := viper.GetString(flagDatabaseName)
	collectionNodes := viper.GetString(flagCollectionNodes)
	collectionSecrets := viper.GetString(flagCollectionSecrets)

	if collectionNodes == "" {
		collectionNodes = "nodes"
	}
	if collectionSecrets == "" {
		collectionSecrets = "secrets"
	}

	if strings.HasPrefix(databaseURL, "memongodb://") {
		server := mongoserver.New()
		defer server.Stop()

		databaseURL = server.URI()
	}

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

		collection := client.Database(databaseName).Collection(collectionNodes)
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
