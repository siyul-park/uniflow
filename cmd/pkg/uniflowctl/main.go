package main

import (
	"context"
	"github.com/siyul-park/uniflow/cmd/pkg/driver"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/siyul-park/uniflow/cmd/pkg/cli"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const configFile = ".uniflow.toml"

const (
	flagDatabaseURL       = "database.url"
	flagDatabaseName      = "database.name"
	flagCollectionSpecs   = "collection.specs"
	flagCollectionSecrets = "collection.secrets"
	flagCollectionCharts  = "collection.charts"
)

func init() {
	viper.SetDefault(flagCollectionSpecs, "specs")
	viper.SetDefault(flagCollectionSecrets, "secrets")
	viper.SetDefault(flagCollectionCharts, "charts")

	viper.SetConfigFile(configFile)
	viper.AutomaticEnv()
	viper.ReadInConfig()
}

func main() {
	ctx := context.Background()

	databaseURL := viper.GetString(flagDatabaseURL)
	databaseName := viper.GetString(flagDatabaseName)
	collectionNodes := viper.GetString(flagCollectionSpecs)
	collectionSecrets := viper.GetString(flagCollectionSecrets)
	collectionCharts := viper.GetString(flagCollectionCharts)

	d := driver.NewInMemoryDriver()
	defer d.Close(ctx)

	if strings.HasPrefix(databaseURL, "memongodb://") || strings.HasPrefix(databaseURL, "mongodb://") {
		var err error
		if d, err = driver.NewMongoDriver(ctx, databaseURL, databaseName); err != nil {
			log.Fatal(err)
		}
	}

	specStore, err := d.SpecStore(ctx, collectionNodes)
	if err != nil {
		log.Fatal(err)
	}
	secretStore, err := d.SecretStore(ctx, collectionSecrets)
	if err != nil {
		log.Fatal(err)
	}
	chartStore, err := d.ChartStore(ctx, collectionCharts)
	if err != nil {
		log.Fatal(err)
	}

	fs := afero.NewOsFs()

	cmd := cli.NewCommand(cli.Config{
		Use:   "uniflowctl",
		Short: "A high-performance, extremely flexible, and easily extensible universal workflow engine.",
	})
	cmd.AddCommand(cli.NewApplyCommand(cli.ApplyConfig{
		SpecStore:   specStore,
		SecretStore: secretStore,
		ChartStore:  chartStore,
		FS:          fs,
	}))
	cmd.AddCommand(cli.NewDeleteCommand(cli.DeleteConfig{
		SpecStore:   specStore,
		SecretStore: secretStore,
		ChartStore:  chartStore,
		FS:          fs,
	}))
	cmd.AddCommand(cli.NewGetCommand(cli.GetConfig{
		SpecStore:   specStore,
		SecretStore: secretStore,
		ChartStore:  chartStore,
	}))

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
