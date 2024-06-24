package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/siyul-park/uniflow/cmd/cli"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/database/mongodb"
	"github.com/siyul-park/uniflow/pkg/event"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/siyul-park/uniflow/plugin/pkg/control"
	"github.com/siyul-park/uniflow/plugin/pkg/datastore"
	"github.com/siyul-park/uniflow/plugin/pkg/language"
	"github.com/siyul-park/uniflow/plugin/pkg/language/expr"
	"github.com/siyul-park/uniflow/plugin/pkg/language/javascript"
	"github.com/siyul-park/uniflow/plugin/pkg/language/json"
	"github.com/siyul-park/uniflow/plugin/pkg/language/jsonata"
	"github.com/siyul-park/uniflow/plugin/pkg/language/text"
	"github.com/siyul-park/uniflow/plugin/pkg/language/typescript"
	"github.com/siyul-park/uniflow/plugin/pkg/language/yaml"
	"github.com/siyul-park/uniflow/plugin/pkg/network"
	"github.com/siyul-park/uniflow/plugin/pkg/system"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const configFile = ".uniflow.toml"

const (
	flagDatabaseURL  = "database.url"
	flagDatabaseName = "database.name"
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

	sb := scheme.NewBuilder()
	hb := hook.NewBuilder()

	native := system.NewNativeModule()

	lang := language.NewModule()
	lang.Store(text.Kind, text.NewCompiler())
	lang.Store(json.Kind, json.NewCompiler())
	lang.Store(yaml.Kind, yaml.NewCompiler())
	lang.Store(jsonata.Kind, jsonata.NewCompiler())
	lang.Store(javascript.Kind, javascript.NewCompiler(api.TransformOptions{}))
	lang.Store(typescript.Kind, typescript.NewCompiler())
	lang.Store(expr.Kind, expr.NewCompiler())

	broker := event.NewBroker()
	defer broker.Close()

	sb.Register(control.AddToScheme(control.Config{
		Broker:     broker,
		Module:     lang,
		Expression: expr.Kind,
	}))
	sb.Register(datastore.AddToScheme())
	sb.Register(network.AddToScheme())
	sb.Register(system.AddToScheme(native))

	hb.Register(control.AddToHook(control.Config{
		Broker:     broker,
		Module:     lang,
		Expression: expr.Kind,
	}))
	hb.Register(network.AddToHook())

	sc, err := sb.Build()
	if err != nil {
		log.Fatal(err)
	}
	hk, err := hb.Build()
	if err != nil {
		log.Fatal(err)
	}

	var db database.Database
	if strings.HasPrefix(databaseURL, "mongodb://") {
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		opts := options.Client().ApplyURI(databaseURL).SetServerAPIOptions(serverAPI)
		client, err := mongodb.Connect(ctx, opts)
		if err != nil {
			log.Fatal(err)
		}
		db, err = client.Database(ctx, databaseName)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		db = memdb.New(databaseName)
	}

	st, err := store.New(ctx, store.Config{
		Scheme:   sc,
		Database: db,
	})
	if err != nil {
		log.Fatal(err)
	}

	native.Store(system.OPCreateNodes, system.CreateNodes(st))
	native.Store(system.OPReadNodes, system.ReadNodes(st))
	native.Store(system.OPUpdateNodes, system.UpdateNodes(st))
	native.Store(system.OPDeleteNodes, system.DeleteNodes(st))

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fsys := os.DirFS(wd)

	cmd := cli.NewCommand(cli.Config{
		Scheme:   sc,
		Hook:     hk,
		Database: db,
		FS:       fsys,
	})

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
