package runtime

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/database"
	"github.com/siyul-park/uniflow/database/memdb"
	"github.com/siyul-park/uniflow/hook"
	"github.com/siyul-park/uniflow/loader"
	"github.com/siyul-park/uniflow/scheme"
	"github.com/siyul-park/uniflow/store"
	"github.com/siyul-park/uniflow/symbol"
)

// Config holds the configuration options for the Runtime.
type Config struct {
	Namespace string
	Hook      *hook.Hook
	Scheme    *scheme.Scheme
	Database  database.Database
}

// Runtime represents an execution environment for running Flows.
type Runtime struct {
	store      *store.Store
	table      *symbol.Table
	loader     *loader.Loader
	reconciler *loader.Reconciler
}

// New creates a new Runtime instance with the specified configuration.
func New(ctx context.Context, config Config) (*Runtime, error) {
	if config.Hook == nil {
		config.Hook = hook.New()
	}
	if config.Scheme == nil {
		config.Scheme = scheme.New()
	}
	if config.Database == nil {
		config.Database = memdb.New("")
	}

	st, err := store.New(ctx, store.Config{
		Scheme:   config.Scheme,
		Database: config.Database,
	})
	if err != nil {
		return nil, err
	}

	tb := symbol.NewTable(config.Scheme, symbol.TableOptions{
		LoadHooks:   []symbol.LoadHook{config.Hook},
		UnloadHooks: []symbol.UnloadHook{config.Hook},
	})

	ld := loader.New(loader.Config{
		Namespace: config.Namespace,
		Store:     st,
		Table:     tb,
	})

	rc := loader.NewReconciler(loader.ReconcilerConfig{
		Namespace: config.Namespace,
		Store:     st,
		Loader:    ld,
	})

	return &Runtime{
		store:      st,
		table:      tb,
		loader:     ld,
		reconciler: rc,
	}, nil
}

// Lookup searches for a node.Node in the symbol.Table. If not found, it loads it from store.Store.
func (r *Runtime) Lookup(ctx context.Context, id uuid.UUID) (*symbol.Symbol, error) {
	if s, ok := r.table.LookupByID(id); !ok {
		return r.loader.LoadOne(ctx, id)
	} else {
		return s, nil
	}
}

// Start initiates the Runtime.
// It loads all spec.Specs as node.Nodes from the database.Collection,
// and continuously monitors and runs them by staying up-to-date with spec.Spec changes.
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
