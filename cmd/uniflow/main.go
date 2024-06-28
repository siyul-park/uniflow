package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/siyul-park/uniflow/cmd/cli"
	"github.com/siyul-park/uniflow/database"
	"github.com/siyul-park/uniflow/database/memdb"
	"github.com/siyul-park/uniflow/database/mongodb"
	"github.com/siyul-park/uniflow/event"
	ctrlx "github.com/siyul-park/uniflow/ext/ctrl"
	eventx "github.com/siyul-park/uniflow/ext/event"
	iox "github.com/siyul-park/uniflow/ext/io"
	"github.com/siyul-park/uniflow/ext/language"
	"github.com/siyul-park/uniflow/ext/language/cel"
	"github.com/siyul-park/uniflow/ext/language/javascript"
	"github.com/siyul-park/uniflow/ext/language/json"
	"github.com/siyul-park/uniflow/ext/language/text"
	"github.com/siyul-park/uniflow/ext/language/typescript"
	"github.com/siyul-park/uniflow/ext/language/yaml"
	netx "github.com/siyul-park/uniflow/ext/net"
	sysx "github.com/siyul-park/uniflow/ext/sys"
	"github.com/siyul-park/uniflow/hook"
	"github.com/siyul-park/uniflow/scheme"
	"github.com/siyul-park/uniflow/store"
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

	natives := sysx.NewNativeModule()

	langs := language.NewModule()
	langs.Store(text.Kind, text.NewCompiler())
	langs.Store(json.Kind, json.NewCompiler())
	langs.Store(yaml.Kind, yaml.NewCompiler())
	langs.Store(cel.Kind, cel.NewCompiler())
	langs.Store(javascript.Kind, javascript.NewCompiler())
	langs.Store(typescript.Kind, typescript.NewCompiler())

	broker := event.NewBroker()
	defer broker.Close()

	sb.Register(ctrlx.AddToScheme(langs, cel.Kind))
	sb.Register(eventx.AddToScheme(broker))
	sb.Register(iox.AddToScheme())
	sb.Register(netx.AddToScheme())
	sb.Register(sysx.AddToScheme(natives))

	hb.Register(eventx.AddToHook(broker))
	hb.Register(netx.AddToHook())

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

	natives.Store(sysx.OPCreateNodes, sysx.CreateNodes(st))
	natives.Store(sysx.OPReadNodes, sysx.ReadNodes(st))
	natives.Store(sysx.OPUpdateNodes, sysx.UpdateNodes(st))
	natives.Store(sysx.OPDeleteNodes, sysx.DeleteNodes(st))

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
