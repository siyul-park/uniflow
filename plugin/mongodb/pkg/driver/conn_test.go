package driver

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/siyul-park/uniflow/plugin/mongodb/internal/server"
)

func TestConn_Load(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	srv := server.New()
	defer server.Release(srv)

	con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
	defer con.Disconnect(ctx)

	c := newConn(con, faker.UUIDHyphenated())
	defer c.Close()

	s, err := c.Load(faker.UUIDHyphenated())
	require.NoError(t, err)
	require.NotNil(t, s)
}
