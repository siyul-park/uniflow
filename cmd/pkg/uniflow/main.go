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

	sbuilder := scheme.NewBuilder()
	hbuilder := hook.NewBuilder()

	langs := language.NewModule()
	langs.Store(text.Language, text.NewCompiler())
	langs.Store(json.Language, json.NewCompiler())
	langs.Store(yaml.Language, yaml.NewCompiler())
	langs.Store(cel.Language, cel.NewCompiler())
	langs.Store(javascript.Language, javascript.NewCompiler())
	langs.Store(typescript.Language, typescript.NewCompiler())

	stable := system.NewTable()
	stable.Store(system.CodeCreateNodes, system.CreateNodes(specStore))
	stable.Store(system.CodeReadNodes, system.ReadNodes(specStore))
	stable.Store(system.CodeUpdateNodes, system.UpdateNodes(specStore))
	stable.Store(system.CodeDeleteNodes, system.DeleteNodes(specStore))

	sbuilder.Register(control.AddToScheme(langs, cel.Language))
	sbuilder.Register(io.AddToScheme())
	sbuilder.Register(network.AddToScheme())
	sbuilder.Register(system.AddToScheme(stable))

	hbuilder.Register(control.AddToHook())
	hbuilder.Register(network.AddToHook())

	scheme, err := sbuilder.Build()
	if err != nil {
		log.Fatal(err)
	}
	hook, err := hbuilder.Build()
	if err != nil {
		log.Fatal(err)
	}

	fsys := afero.NewOsFs()

	cmd := cli.NewCommand(cli.Config{
		Scheme:      scheme,
		Hook:        hook,
		SpecStore:   specStore,
		SecretStore: secretStore,
		FS:          fsys,
	})

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
