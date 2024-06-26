package runtime

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/loader"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/siyul-park/uniflow/pkg/symbol"
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

// Insert inserts a spec.Spec into the Runtime and returns the corresponding symbol.Symbol.
func (r *Runtime) Insert(ctx context.Context, spc spec.Spec) (*symbol.Symbol, error) {
	if _, err := r.store.InsertOne(ctx, spc); err != nil {
		return nil, err
	} else {
		return r.table.Insert(spc)
	}
}

// Free removes a spec.Spec from the Runtime and returns whether it was successfully deleted.
func (r *Runtime) Free(ctx context.Context, spc spec.Spec) (bool, error) {
	if ok, err := r.store.DeleteOne(ctx, store.Where[uuid.UUID](spec.KeyID).EQ(spc.GetID())); err != nil {
		return false, err
	} else if _, err := r.table.Free(spc.GetID()); err != nil {
		return false, err
	} else {
		return ok, nil
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
