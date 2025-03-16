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
	"github.com/siyul-park/uniflow/ext/pkg/control"
	"github.com/siyul-park/uniflow/ext/pkg/io"
	"github.com/siyul-park/uniflow/ext/pkg/language"
	"github.com/siyul-park/uniflow/ext/pkg/language/cel"
	"github.com/siyul-park/uniflow/ext/pkg/language/javascript"
	"github.com/siyul-park/uniflow/ext/pkg/language/json"
	"github.com/siyul-park/uniflow/ext/pkg/language/text"
	"github.com/siyul-park/uniflow/ext/pkg/language/typescript"
	"github.com/siyul-park/uniflow/ext/pkg/language/yaml"
	"github.com/siyul-park/uniflow/ext/pkg/network"
	"github.com/siyul-park/uniflow/ext/pkg/testing"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/store"
	testing2 "github.com/siyul-park/uniflow/pkg/testing"
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

	runner := testing2.NewRunner()

	schemeBuilder := scheme.NewBuilder()
	hookBuilder := hook.NewBuilder()

	languages := language.NewModule()
	languages.Store(text.Language, text.NewCompiler())
	languages.Store(json.Language, json.NewCompiler())
	languages.Store(yaml.Language, yaml.NewCompiler())
	languages.Store(cel.Language, cel.NewCompiler())
	languages.Store(javascript.Language, javascript.NewCompiler())
	languages.Store(typescript.Language, typescript.NewCompiler())

	schemeBuilder.Register(control.AddToScheme(languages, cel.Language))
	schemeBuilder.Register(io.AddToScheme(io.NewOSFileSystem()))
	schemeBuilder.Register(network.AddToScheme())
	schemeBuilder.Register(testing.AddToScheme())

	hookBuilder.Register(network.AddToHook())
	hookBuilder.Register(testing.AddToHook(runner))

	scheme, err := schemeBuilder.Build()
	if err != nil {
		log.Fatal(err)
	}
	hook, err := hookBuilder.Build()
	if err != nil {
		log.Fatal(err)
	}

	fs := afero.NewOsFs()

	cmd := cli.NewCommand(cli.Config{
		Use:   "uniflow",
		Short: "A high-performance, extremely flexible, and easily extensible universal workflow engine.",
		FS:    fs,
	})
	cmd.AddCommand(cli.NewStartCommand(cli.StartConfig{
		Scheme:     scheme,
		Hook:       hook,
		SpecStore:  specStore,
		ValueStore: valueStore,
		FS:         fs,
	}))
	cmd.AddCommand(cli.NewTestCommand(cli.TestConfig{
		Runner:     runner,
		Scheme:     scheme,
		Hook:       hook,
		SpecStore:  specStore,
		ValueStore: valueStore,
		FS:         fs,
	}))

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
