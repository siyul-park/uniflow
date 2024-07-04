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
	namespace string
	store     *store.Store
	table     *symbol.Table
	loader    *loader.Loader
}

// New creates a new Runtime instance with the specified configuration.
func New(ctx context.Context, config Config) (*Runtime, error) {
	if config.Namespace == "" {
		config.Namespace = spec.DefaultNamespace
	}
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

	return &Runtime{
		namespace: config.Namespace,
		store:     st,
		table:     tb,
		loader:    ld,
	}, nil
}

// LookupByID retrieves a node from the table or loads it from the store if not found.
func (r *Runtime) LookupByID(ctx context.Context, id uuid.UUID) (*symbol.Symbol, error) {
	if s, ok := r.table.LookupByID(id); ok {
		return s, nil
	}
	return r.loader.LoadOne(ctx, id)
}

// LookupByName retrieves a symbol by name from the table or loads it from the store if not found.
func (r *Runtime) LookupByName(ctx context.Context, name string) (*symbol.Symbol, error) {
	if s, ok := r.table.LookupByName(r.namespace, name); ok {
		return s, nil
	}

	filter := store.Where[string](spec.KeyNamespace).EQ(r.namespace).
		And(store.Where[string](spec.KeyName).EQ(name))

	s, err := r.store.FindOne(ctx, filter)
	if err != nil || s == nil {
		return nil, err
	}

	return r.LookupByID(ctx, s.GetID())
}

// Insert adds a spec to the Runtime and returns the corresponding symbol.
func (r *Runtime) Insert(ctx context.Context, spc spec.Spec) (*symbol.Symbol, error) {
	if _, err := r.store.InsertOne(ctx, spc); err != nil {
		return nil, err
	}
	return r.table.Insert(spc)
}

// Free removes a spec from the Runtime and returns whether it was successfully deleted.
func (r *Runtime) Free(ctx context.Context, spc spec.Spec) (bool, error) {
	ok, err := r.store.DeleteOne(ctx, store.Where[uuid.UUID](spec.KeyID).EQ(spc.GetID()))
	if err != nil {
		return false, err
	}
	if _, err := r.table.Free(spc.GetID()); err != nil {
		return false, err
	}
	return ok, nil
}

// Load loads all symbols from the store.
func (r *Runtime) Load(ctx context.Context) ([]*symbol.Symbol, error) {
	return r.loader.LoadAll(ctx)
}

// Watch starts the reconciler's watch process.
func (r *Runtime) Watch(ctx context.Context) error {
	return r.loader.Watch(ctx)
}

// Start begins the reconciliation process of the Runtime.
func (r *Runtime) Start(ctx context.Context) error {
	return r.loader.Reconcile(ctx)
}

// Close shuts down the Runtime by closing the loader and clearing the symbol table.
func (r *Runtime) Close() error {
	if err := r.loader.Close(); err != nil {
		return err
	}
	return r.table.Clear()
}
