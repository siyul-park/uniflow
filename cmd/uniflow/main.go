package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/siyul-park/uniflow/cmd/uniflow/uniflow"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/database/mongodb"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/siyul-park/uniflow/third_party/controllx"
	"github.com/siyul-park/uniflow/third_party/networkx"
	"github.com/siyul-park/uniflow/third_party/systemx"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	configFile = ".uniflow.toml"
)

func init() {
	viper.SetConfigFile(configFile)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func main() {
	if err := execute(); err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
}

func execute() error {
	ctx := context.Background()

	sb := scheme.NewBuilder(
		controllx.AddToScheme(),
		networkx.AddToScheme(),
	)
	hb := hook.NewBuilder(
		networkx.AddToHooks(),
	)

	sc, err := sb.Build()
	if err != nil {
		return err
	}
	hk, err := hb.Build()
	if err != nil {
		return err
	}

	db, err := loadDB(ctx)
	if err != nil {
		return err
	}

	curDir, err := os.Getwd()
	if err != nil {
		return err
	}
	fsys := os.DirFS(curDir)

	st, err := storage.New(ctx, storage.Config{
		Scheme:   sc,
		Database: db,
	})
	if err != nil {
		return err
	}
	systemx.AddToScheme(st)(sc)

	cmd := uniflow.NewCmd(uniflow.Config{
		Scheme:   sc,
		Hook:     hk,
		Database: db,
		FS:       fsys,
	})
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

func loadDB(ctx context.Context) (database.Database, error) {
	dbURL := viper.GetString(FlagDatabaseURL)
	dbName := viper.GetString(FlagDatabaseName)

	if dbURL == "" || strings.HasPrefix(dbURL, "mem://") {
		return memdb.New(dbName), nil
	} else if strings.HasPrefix(dbURL, "mongodb://") {
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		opts := options.Client().ApplyURI(dbURL).SetServerAPIOptions(serverAPI)
		client, err := mongo.Connect(ctx, opts)
		if err != nil {
			return nil, err
		}
		return mongodb.NewDatabase(client.Database(dbName)), nil
	}
	return nil, fmt.Errorf("%s is invalid", FlagDatabaseURL)
}
