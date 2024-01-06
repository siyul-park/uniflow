package runtime

import (
	"context"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/loader"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Config holds the configuration options for the Runtime.
type Config struct {
	Namespace string
	Hooks     *hook.Hook
	Scheme    *scheme.Scheme
	Database  database.Database
}

// Runtime represents an execution environment for running Flows.
type Runtime struct {
	storage    *storage.Storage
	table      *symbol.Table
	loader     *loader.Loader
	reconciler *loader.Reconciler
}

// New creates a new Runtime instance with the specified configuration.
func New(ctx context.Context, config Config) (*Runtime, error) {
	if config.Hooks == nil {
		config.Hooks = hook.New()
	}
	if config.Scheme == nil {
		config.Scheme = scheme.New()
	}
	if config.Database == nil {
		config.Database = memdb.New("")
	}

	st, err := storage.New(ctx, storage.Config{
		Scheme:   config.Scheme,
		Database: config.Database,
	})
	if err != nil {
		return nil, err
	}

	tb := symbol.NewTable(config.Scheme, symbol.TableOptions{
		LoadHooks:   []symbol.LoadHook{config.Hooks},
		UnloadHooks: []symbol.UnloadHook{config.Hooks},
	})

	ld := loader.New(loader.Config{
		Namespace: config.Namespace,
		Storage:   st,
		Table:     tb,
	})

	rc := loader.NewReconciler(loader.ReconcilerConfig{
		Namespace: config.Namespace,
		Storage:   st,
		Loader:    ld,
	})

	return &Runtime{
		storage:    st,
		table:      tb,
		loader:     ld,
		reconciler: rc,
	}, nil
}

// Lookup searches for a node.Node in the symbol.Table. If not found, it loads it from storage.Storage.
func (r *Runtime) Lookup(ctx context.Context, id ulid.ULID) (*symbol.Symbol, error) {
	if s, ok := r.table.LookupByID(id); !ok {
		return r.loader.LoadOne(ctx, id)
	} else {
		return s, nil
	}
}

// Start initiates the Runtime.
// It loads all scheme.Specs as node.Nodes from the database.Collection,
// and continuously monitors and runs them by staying up-to-date with scheme.Spec changes.
func (r *Runtime) Start(ctx context.Context) error {
	if err := r.reconciler.Watch(ctx); err != nil {
		return err
	}
	if _, err := r.loader.LoadAll(ctx); err != nil {
		return err
	}
	return r.reconciler.Reconcile(ctx)
}

// Close shuts down the Runtime.
func (r *Runtime) Close() error {
	if err := r.reconciler.Close(); err != nil {
		return err
	}
	return r.table.Clear()
}
