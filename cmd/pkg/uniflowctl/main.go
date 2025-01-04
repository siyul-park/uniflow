package main

import (
	"context"
	"log"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/siyul-park/uniflow/cmd/pkg/cli"
	"github.com/siyul-park/uniflow/cmd/pkg/driver"
	"github.com/spf13/afero"
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

	_ = k.Load(file.Provider(configFile), toml.Parser())

	if err := k.Load(env.Provider("", ".", func(s string) string {
		return strcase.ToDelimited(s, '.')
	}), nil); err != nil {
		log.Fatal(err)
	}
}

func main() {
	ctx := context.Background()

	databaseURL := k.String(cli.EnvDatabaseURL)
	databaseName := k.String(cli.EnvDatabaseName)
	collectionNodes := k.String(cli.EnvCollectionSpecs)
	collectionValues := k.String(cli.EnvCollectionValues)

	drv := driver.NewInMemoryDriver()
	defer drv.Close(ctx)

	if strings.HasPrefix(databaseURL, "memongodb://") || strings.HasPrefix(databaseURL, "mongodb://") {
		var err error
		if drv, err = driver.NewMongoDriver(databaseURL, databaseName); err != nil {
			log.Fatal(err)
		}
	}

	specStore, err := drv.NewSpecStore(ctx, collectionNodes)
	if err != nil {
		log.Fatal(err)
	}
	valueStore, err := drv.NewValueStore(ctx, collectionValues)
	if err != nil {
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
