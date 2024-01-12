package main

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/database/mongodb"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestConnectDatabase(t *testing.T) {
	t.Run("memdb", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		db, err := connectDatabase(ctx, "memdb://", faker.UUIDHyphenated())
		assert.NoError(t, err)
		assert.NotNil(t, db)
	})

	t.Run("mongodb", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		server := mongodb.Server()
		defer mongodb.ReleaseServer(server)

		db, err := connectDatabase(ctx, options.Client().ApplyURI(server.URI()).GetURI(), faker.UUIDHyphenated())
		assert.NoError(t, err)
		assert.NotNil(t, db)
	})
}
