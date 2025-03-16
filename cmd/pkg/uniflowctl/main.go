package main

import (
	"context"
	"log"
	"os/signal"
	"strings"
	"syscall"

	"github.com/iancoleman/strcase"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/siyul-park/uniflow/cmd/pkg/cli"
	mongoserver "github.com/siyul-park/uniflow/driver/mongo/pkg/server"
	mongostore "github.com/siyul-park/uniflow/driver/mongo/pkg/store"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/spf13/afero"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const configFile = ".uniflow.toml"

var k = koanf.New(".")

func init() {
	if err := k.Set(cli.EnvCollectionSpecs, "specs"); err != nil {
		log.Fatal(err)
	}
	if err := k.Set(cli.EnvCollectionValues, "values"); err != nil {
		log.Fatal(err)
	}

	if err := k.Load(file.Provider(configFile), toml.Parser()); err != nil {
		log.Fatal(err)
	}
	if err := k.Load(env.Provider("", ".", func(s string) string {
		return strcase.ToDelimited(s, '.')
	}), nil); err != nil {
		log.Fatal(err)
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	databaseURL := k.String(cli.EnvDatabaseURL)
	databaseName := k.String(cli.EnvDatabaseName)
	collectionNodes := k.String(cli.EnvCollectionSpecs)
	collectionValues := k.String(cli.EnvCollectionValues)

	var source store.Source
	if strings.HasPrefix(databaseURL, "memongodb://") {
		srv := mongoserver.New()
		defer mongoserver.Release(srv)

		client, err := mongo.Connect(
			options.Client().
				ApplyURI(srv.URI()).
				SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1)),
		)
		if err != nil {
			log.Fatal(err)
		}

		database := client.Database(databaseName)
		source = mongostore.NewSource(database)
	} else if strings.HasPrefix(databaseURL, "mongodb://") {
		client, err := mongo.Connect(
			options.Client().
				ApplyURI(databaseURL).
				SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1)),
		)
		if err != nil {
			log.Fatal(err)
		}

		database := client.Database(databaseName)
		source = mongostore.NewSource(database)
	} else {
		source = store.NewSource()
	}
	defer source.Close()

	specStore, err := source.Open(collectionNodes)
	if err != nil {
		log.Fatal(err)
	}
	valueStore, err := source.Open(collectionValues)
	if err != nil {
		log.Fatal(err)
	}

	if err := specStore.Index(ctx, []string{spec.KeyNamespace, spec.KeyName}, store.IndexOptions{
		Unique: true,
		Filter: map[string]any{spec.KeyName: map[string]any{"$exists": true}},
	}); err != nil {
		log.Fatal(err)
	}
	if err := valueStore.Index(ctx, []string{value.KeyNamespace, value.KeyName}, store.IndexOptions{
		Unique: true,
		Filter: map[string]any{value.KeyName: map[string]any{"$exists": true}},
	}); err != nil {
		log.Fatal(err)
	}

	fs := afero.NewOsFs()

	cmd := cli.NewCommand(cli.Config{
		Use:   "uniflowctl",
		Short: "A high-performance, extremely flexible, and easily extensible universal workflow engine.",
	})
	cmd.AddCommand(cli.NewApplyCommand(cli.ApplyConfig{
		SpecStore:  specStore,
		ValueStore: valueStore,
		FS:         fs,
	}))
	cmd.AddCommand(cli.NewDeleteCommand(cli.DeleteConfig{
		SpecStore:  specStore,
		ValueStore: valueStore,
		FS:         fs,
	}))
	cmd.AddCommand(cli.NewGetCommand(cli.GetConfig{
		SpecStore:  specStore,
		ValueStore: valueStore,
	}))

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
