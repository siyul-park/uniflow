package main

import (
	"context"
	"net/url"
	"os/signal"
	"syscall"

	"github.com/spf13/afero"

	"github.com/siyul-park/uniflow/internal/cli"
	"github.com/siyul-park/uniflow/pkg/driver"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/plugin"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/testing"
	"github.com/siyul-park/uniflow/pkg/value"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	runner := testing.NewRunner()

	schemeBuilder := scheme.NewBuilder()
	hookBuilder := hook.NewBuilder()

	driverRegistry := driver.NewRegistry()
	defer driverRegistry.Close()

	fatal(driverRegistry.Register("memory", driver.New()))

	pluginRegistry := plugin.NewRegistry()
	defer pluginRegistry.Unload(ctx)

	for _, cfg := range config.Slices(envPlugin) {
		p := must(plugin.Open(cfg.String(keyPath), cfg.Get(keyManifest)))
		fatal(pluginRegistry.Register(p))
	}
	fatal(pluginRegistry.Inject(schemeBuilder, hookBuilder, driverRegistry, runner))
	fatal(pluginRegistry.Load(ctx))

	scheme := must(schemeBuilder.Build())
	hook := must(hookBuilder.Build())

	dsn := must(url.Parse(config.String(envDatabaseURL)))
	drv := must(driverRegistry.Lookup(dsn.Scheme))
	conn := must(drv.Open(dsn.String()))

	specStore := must(conn.Load(config.String(envCollectionSpecs)))
	valueStore := must(conn.Load(config.String(envCollectionValues)))

	fatal(specStore.Index(ctx, []string{spec.KeyNamespace, spec.KeyName}, driver.IndexOptions{
		Unique: true,
		Filter: map[string]any{spec.KeyName: map[string]any{"$exists": true}},
	}))
	fatal(valueStore.Index(ctx, []string{value.KeyNamespace, value.KeyName}, driver.IndexOptions{
		Unique: true,
		Filter: map[string]any{value.KeyName: map[string]any{"$exists": true}},
	}))

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

	fatal(cmd.Execute())
}
