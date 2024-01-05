package main

import (
	"context"
	"strings"

	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/database/mongodb"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func connectDatabase(ctx context.Context, url, name string) (database.Database, error) {
	if strings.HasPrefix(url, "mongodb://") {
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		opts := options.Client().ApplyURI(url).SetServerAPIOptions(serverAPI)
		client, err := mongodb.Connect(ctx, opts)
		if err != nil {
			return nil, err
		}
		return client.Database(ctx, name)
	}
	return memdb.New(name), nil
}
