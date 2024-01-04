package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/database/mongodb"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	flagDatabaseURL  = "database.url"
	flagDatabaseName = "database.name"
)

func connectDatabase(ctx context.Context) (database.Database, error) {
	dbURL := viper.GetString(flagDatabaseURL)
	dbName := viper.GetString(flagDatabaseName)

	if dbURL == "" || strings.HasPrefix(dbURL, "mem://") {
		return memdb.New(dbName), nil
	} else if strings.HasPrefix(dbURL, "mongodb://") {
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		opts := options.Client().ApplyURI(dbURL).SetServerAPIOptions(serverAPI)
		client, err := mongodb.Connect(ctx, opts)
		if err != nil {
			return nil, err
		}
		return client.Database(ctx, dbName)
	}
	return nil, fmt.Errorf("%s is invalid", flagDatabaseURL)
}
