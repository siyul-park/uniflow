package main

import (
	"github.com/iancoleman/strcase"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const (
	envDatabaseURL      = "database.url"
	envCollectionSpecs  = "collection.specs"
	envCollectionValues = "collection.values"
	envPlugin           = "plugin"

	keyPath     = "path"
	keyManifest = "manifest"
)

const configFile = ".uniflow.toml"

var config = koanf.New(".")

func init() {
	fatal(config.Set(envDatabaseURL, "memory://"))
	fatal(config.Set(envCollectionSpecs, "specs"))
	fatal(config.Set(envCollectionValues, "values"))

	_ = config.Load(file.Provider(configFile), toml.Parser())

	fatal(config.Load(env.Provider("", ".", func(s string) string {
		return strcase.ToDelimited(s, '.')
	}), nil))
}
