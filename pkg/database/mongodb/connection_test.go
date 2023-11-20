package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	con, err := Connect(ctx, server.URI())
	assert.NoError(t, err)
	assert.NotNil(t, con)
}

func TestConnection_Database(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	con, _ := Connect(ctx, server.URI())

	dbname := faker.UUIDHyphenated()

	db, err := con.Database(ctx, dbname)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	assert.Equal(t, dbname, db.Name())
}

func TestConnection_Disconnect(t *testing.T) {
	server := Server()
	defer ReleaseServer(server)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	con, _ := Connect(ctx, server.URI())

	err := con.Disconnect(ctx)
	assert.NoError(t, err)
}
