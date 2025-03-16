package store

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/driver/mongo/pkg/server"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestSource_Open(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	srv := server.New()
	defer server.Release(srv)

	con, _ := mongo.Connect(options.Client().ApplyURI(srv.URI()))
	defer con.Disconnect(ctx)

	src := NewSource(con.Database(faker.UUIDHyphenated()))
	defer src.Close()

	name := faker.UUIDHyphenated()

	s1, err := src.Open(name)
	require.NoError(t, err)

	s2, err := src.Open(name)
	require.NoError(t, err)
	require.Equal(t, s1, s2)
}
