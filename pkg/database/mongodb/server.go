package mongodb

import (
	"context"
	"sync"
	"time"

	"github.com/tryvium-travels/memongo"
	"github.com/tryvium-travels/memongo/memongolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	serverStartUpLock sync.Mutex
	serverPool        = sync.Pool{
		New: func() any {
			serverStartUpLock.Lock()
			defer serverStartUpLock.Unlock()

			opts := &memongo.Options{
				MongoVersion:     "6.0.8",
				LogLevel:         memongolog.LogLevelWarn,
				ShouldUseReplica: true,
			}

			if server, err := memongo.StartWithOptions(opts); err == nil {
				return server
			} else {
				panic(err)
			}
		},
	}
)

func Server() *memongo.Server {
	return serverPool.Get().(*memongo.Server)
}

func ReleaseServer(server *memongo.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if client, err := mongo.Connect(ctx, options.Client().ApplyURI(server.URI()+"/retryWrites=false")); err == nil {
		if databases, err := client.ListDatabaseNames(ctx, bson.D{}); err == nil {
			for _, db := range databases {
				_ = client.Database(db).Drop(ctx)
			}
		}
		_ = client.Disconnect(ctx)
		serverPool.Put(server)
		return
	}

	server.Stop()
}
