package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/siyul-park/uniflow/cmd/cli"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/database/mongodb"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/siyul-park/uniflow/plugin/pkg/control"
	"github.com/siyul-park/uniflow/plugin/pkg/datastore"
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

	table := system.NewNativeTable()

	sb.Register(control.AddToScheme())
	sb.Register(datastore.AddToScheme())
	sb.Register(network.AddToScheme())
	sb.Register(system.AddToScheme(table))

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

	st, err := storage.New(ctx, storage.Config{
		Scheme:   sc,
		Database: db,
	})
	if err != nil {
		log.Fatal(err)
	}

	table.Store(system.OPCreateNodes, system.CreateNodes(st))
	table.Store(system.OPReadNodes, system.ReadNodes(st))
	table.Store(system.OPUpdateNodes, system.UpdateNodes(st))
	table.Store(system.OPDeleteNodes, system.DeleteNodes(st))

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
