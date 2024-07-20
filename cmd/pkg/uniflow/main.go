package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/siyul-park/uniflow/cmd/pkg/cli"
	_ "github.com/siyul-park/uniflow/driver/mongo/pkg/encoding"
	mongoserver "github.com/siyul-park/uniflow/driver/mongo/pkg/server"
	mongospec "github.com/siyul-park/uniflow/driver/mongo/pkg/spec"
	"github.com/siyul-park/uniflow/ext/pkg/control"
	"github.com/siyul-park/uniflow/ext/pkg/event"
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
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const configFile = ".uniflow.toml"

const (
	flagDatabaseURL     = "database.url"
	flagDatabaseName    = "database.name"
	flagCollectionNodes = "collection.nodes"
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

	if collectionNodes == "" {
		collectionNodes = "nodes"
	}

	if strings.HasPrefix(databaseURL, "memongodb://") {
		server := mongoserver.New()
		defer mongoserver.Release(server)

		databaseURL = server.URI()
	}

	var store spec.Store
	if strings.HasPrefix(databaseURL, "mongodb://") {
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		opts := options.Client().ApplyURI(databaseURL).SetServerAPIOptions(serverAPI)
		client, err := mongo.Connect(ctx, opts)
		if err != nil {
			log.Fatal(err)
		}
		collection := client.Database(databaseName).Collection(collectionNodes)

		s := mongospec.NewStore(collection)
		if err := s.Index(ctx); err != nil {
			log.Fatal(err)
		}
		store = s
	} else {
		store = spec.NewMemStore()
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
	stable.Store(system.CodeCreateNodes, system.CreateNodes(store))
	stable.Store(system.CodeReadNodes, system.ReadNodes(store))
	stable.Store(system.CodeUpdateNodes, system.UpdateNodes(store))
	stable.Store(system.CodeDeleteNodes, system.DeleteNodes(store))

	broker := event.NewBroker()
	defer broker.Close()

	sbuilder.Register(control.AddToScheme(langs, cel.Language))
	sbuilder.Register(event.AddToScheme(broker, broker))
	sbuilder.Register(io.AddToScheme())
	sbuilder.Register(network.AddToScheme())
	sbuilder.Register(system.AddToScheme(stable))

	hbuilder.Register(control.AddToHook())
	hbuilder.Register(event.AddToHook(broker))
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
		Scheme: scheme,
		Hook:   hook,
		Store:  store,
		FS:     fsys,
	})

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
