package main

import (
	"context"
	"log"
	"strings"

	"github.com/siyul-park/uniflow/pkg/testing"

	"github.com/iancoleman/strcase"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/siyul-park/uniflow/cmd/pkg/cli"
	"github.com/siyul-park/uniflow/cmd/pkg/driver"
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
	"github.com/siyul-park/uniflow/ext/pkg/system"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/spf13/afero"
)

const configFile = ".uniflow.toml"

const (
	topicSpecs  = "specs"
	topicValues = "values"

	opCreateSpecs = "specs.create"
	opReadSpecs   = "specs.read"
	opUpdateSpecs = "specs.update"
	opDeleteSpecs = "specs.delete"

	opCreateValues = "values.create"
	opReadValues   = "values.read"
	opUpdateValues = "values.update"
	opDeleteValues = "values.delete"
)

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

	runner := testing.NewRunner()

	schemeBuilder := scheme.NewBuilder()
	hookBuilder := hook.NewBuilder()

	languages := language.NewModule()
	languages.Store(text.Language, text.NewCompiler())
	languages.Store(json.Language, json.NewCompiler())
	languages.Store(yaml.Language, yaml.NewCompiler())
	languages.Store(cel.Language, cel.NewCompiler())
	languages.Store(javascript.Language, javascript.NewCompiler())
	languages.Store(typescript.Language, typescript.NewCompiler())

	signals := map[string]any{
		topicSpecs:  system.WatchResource(specStore),
		topicValues: system.WatchResource(valueStore),
	}
	calls := map[string]any{
		opCreateSpecs:  system.CreateResource(specStore),
		opReadSpecs:    system.ReadResource(specStore),
		opUpdateSpecs:  system.UpdateResource(specStore),
		opDeleteSpecs:  system.DeleteResource(specStore),
		opCreateValues: system.CreateResource(valueStore),
		opReadValues:   system.ReadResource(valueStore),
		opUpdateValues: system.UpdateResource(valueStore),
		opDeleteValues: system.DeleteResource(valueStore),
	}

	systemAddToScheme := system.AddToScheme()

	for topic, signal := range signals {
		if err := systemAddToScheme.SetSignal(topic, signal); err != nil {
			log.Fatal(err)
		}
	}
	for opcode, call := range calls {
		if err := systemAddToScheme.SetCall(opcode, call); err != nil {
			log.Fatal(err)
		}
	}

	schemeBuilder.Register(control.AddToScheme(languages, cel.Language))
	schemeBuilder.Register(io.AddToScheme(io.NewOSFileSystem()))
	schemeBuilder.Register(network.AddToScheme())
	schemeBuilder.Register(systemAddToScheme)

	hookBuilder.Register(network.AddToHook())
	hookBuilder.Register(system.AddToHook())

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
