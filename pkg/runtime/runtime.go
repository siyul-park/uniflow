package runtime

import (
	"context"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/loader"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Config holds the configuration options for the Runtime.
type Config struct {
	Namespace string           // Namespace is the namespace for the Runtime.
	Hooks     *hook.Hook       // Hooks represent the hooks for the Runtime.
	Scheme    *scheme.Scheme   // Scheme is the scheme for the Runtime.
	Database  database.Database // Database is the database for the Runtime.
}

// Runtime represents an execution environment for running Flows.
type Runtime struct {
	namespace  string    
	hooks      *hook.Hook   
	scheme     *scheme.Scheme 
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

	tb := symbol.NewTable(symbol.TableOptions{
		LoadHooks:   []symbol.LoadHook{config.Hooks},
		UnloadHooks: []symbol.UnloadHook{config.Hooks},
	})

	ld := loader.New(loader.Config{
		Namespace: config.Namespace,
		Scheme:    config.Scheme,
		Storage:   st,
		Table:     tb,
	})
	if err != nil {
		return nil, err
	}

	var filter *storage.Filter
	if config.Namespace != "" {
		filter = storage.Where[string](scheme.KeyNamespace).EQ(config.Namespace)
	}

	rc := loader.NewReconciler(loader.ReconcilerConfig{
		Storage: st,
		Loader:  ld,
		Filter:  filter,
	})

	return &Runtime{
		namespace:  config.Namespace,
		hooks:      config.Hooks,
		scheme:     config.Scheme,
		storage:    st,
		table:      tb,
		loader:     ld,
		reconciler: rc,
	}, nil
}

// Lookup searches for a node.Node in the symbol.Table. If not found, it loads it from storage.Storage.
func (r *Runtime) Lookup(ctx context.Context, id ulid.ULID) (node.Node, error) {
	if s, ok := r.table.LookupByID(id); !ok {
		return r.loader.LoadOne(ctx, id)
	} else {
		return s, nil
	}
}

// Free unloads a node.Node from the symbol.Table.
func (r *Runtime) Free(_ context.Context, id ulid.ULID) (bool, error) {
	return r.table.Free(id)
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
func (r *Runtime) Close(ctx context.Context) error {
	if err := r.reconciler.Close(); err != nil {
		return err
	}
	return r.table.Close()
}