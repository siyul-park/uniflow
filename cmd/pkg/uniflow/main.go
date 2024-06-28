package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/siyul-park/uniflow/cmd/pkg/cli"
	"github.com/siyul-park/uniflow/extend/pkg/control"
	"github.com/siyul-park/uniflow/extend/pkg/event"
	"github.com/siyul-park/uniflow/extend/pkg/io"
	"github.com/siyul-park/uniflow/extend/pkg/language"
	"github.com/siyul-park/uniflow/extend/pkg/language/cel"
	"github.com/siyul-park/uniflow/extend/pkg/language/javascript"
	"github.com/siyul-park/uniflow/extend/pkg/language/json"
	"github.com/siyul-park/uniflow/extend/pkg/language/text"
	"github.com/siyul-park/uniflow/extend/pkg/language/typescript"
	"github.com/siyul-park/uniflow/extend/pkg/language/yaml"
	"github.com/siyul-park/uniflow/extend/pkg/network"
	"github.com/siyul-park/uniflow/extend/pkg/system"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/database/mongodb"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/store"
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

	natives := system.NewNativeModule()

	langs := language.NewModule()
	langs.Store(text.Kind, text.NewCompiler())
	langs.Store(json.Kind, json.NewCompiler())
	langs.Store(yaml.Kind, yaml.NewCompiler())
	langs.Store(cel.Kind, cel.NewCompiler())
	langs.Store(javascript.Kind, javascript.NewCompiler())
	langs.Store(typescript.Kind, typescript.NewCompiler())

	broker := event.NewBroker()
	defer broker.Close()

	sb.Register(control.AddToScheme(langs, cel.Kind))
	sb.Register(event.AddToScheme(broker, broker))
	sb.Register(io.AddToScheme())
	sb.Register(network.AddToScheme())
	sb.Register(system.AddToScheme(natives))

	hb.Register(event.AddToHook(broker, broker))
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

	natives.Store(system.OPCreateNodes, system.CreateNodes(st))
	natives.Store(system.OPReadNodes, system.ReadNodes(st))
	natives.Store(system.OPUpdateNodes, system.UpdateNodes(st))
	natives.Store(system.OPDeleteNodes, system.DeleteNodes(st))

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
