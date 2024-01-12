package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

const configFile = ".uniflow.toml"

const (
	flagDatabaseURL  = "database.url"
	flagDatabaseName = "database.name"
)

func init() {
	viper.SetConfigFile(configFile)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func main() {
	ctx := context.Background()

	databaseURL := viper.GetString(flagDatabaseURL)
	databaseName := viper.GetString(flagDatabaseName)

	if err := execute(ctx, databaseURL, databaseName); err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
}
