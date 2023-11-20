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

type (
	// Config is a config for for the Runtime.
	Config struct {
		Namespace string
		Hooks     *hook.Hook
		Scheme    *scheme.Scheme
		Database  database.Database
	}

	// Runtime is an execution environment that runs Flows.
	Runtime struct {
		namespace  string
		hooks      *hook.Hook
		scheme     *scheme.Scheme
		storage    *storage.Storage
		table      *symbol.Table
		loader     *loader.Loader
		reconciler *loader.Reconciler
	}
)

// New returns a new Runtime.
func New(ctx context.Context, config Config) (*Runtime, error) {
	ns := config.Namespace
	hk := config.Hooks
	sc := config.Scheme
	db := config.Database

	if hk == nil {
		hk = hook.New()
	}
	if sc == nil {
		sc = scheme.New()
	}
	if db == nil {
		db = memdb.New("")
	}

	st, err := storage.New(ctx, storage.Config{
		Scheme:   sc,
		Database: db,
	})
	if err != nil {
		return nil, err
	}

	tb := symbol.NewTable(symbol.TableOptions{
		PreLoadHooks:    []symbol.PreLoadHook{hk},
		PostLoadHooks:   []symbol.PostLoadHook{hk},
		PreUnloadHooks:  []symbol.PreUnloadHook{hk},
		PostUnloadHooks: []symbol.PostUnloadHook{hk},
	})

	ld, err := loader.New(ctx, loader.Config{
		Scheme:  sc,
		Storage: st,
		Table:   tb,
	})
	if err != nil {
		return nil, err
	}

	var filter *storage.Filter
	if ns != "" {
		filter = storage.Where[string](scheme.KeyNamespace).EQ(ns)
	}
	rc := loader.NewReconciler(loader.ReconcilerConfig{
		Remote: st,
		Loader: ld,
		Filter: filter,
	})

	return &Runtime{
		namespace:  ns,
		hooks:      hk,
		scheme:     sc,
		storage:    st,
		table:      tb,
		loader:     ld,
		reconciler: rc,
	}, nil
}

// Lookup lookup node.Node in symbol.Table, and if it not exist load it from storage.Storage.
func (r *Runtime) Lookup(ctx context.Context, id ulid.ULID) (node.Node, error) {
	filter := storage.Where[ulid.ULID](scheme.KeyID).EQ(id)
	if r.namespace != "" {
		filter = filter.And(storage.Where[string](scheme.KeyNamespace).EQ(r.namespace))
	}
	if s, ok := r.table.Lookup(id); !ok {
		return r.loader.LoadOne(ctx, filter)
	} else {
		return s, nil
	}
}

// Free unload node.Node from symbol.Table.
func (r *Runtime) Free(ctx context.Context, id ulid.ULID) (bool, error) {
	return r.loader.UnloadOne(ctx, storage.Where[ulid.ULID](scheme.KeyID).EQ(id))
}

// Start starts the Runtime.
// Runtime load all scheme.Spec as node.Node from the database.Collection,
// and then keeps node.Node up-to-date and runs by continuously tracking scheme.Spec.
func (r *Runtime) Start(ctx context.Context) error {
	if err := r.reconciler.Watch(ctx); err != nil {
		return err
	}
	var filter *storage.Filter
	if r.namespace != "" {
		filter = filter.And(storage.Where[string](scheme.KeyNamespace).EQ(r.namespace))
	}
	if _, err := r.loader.LoadMany(ctx, filter); err != nil {
		return err
	}
	return r.reconciler.Reconcile(ctx)
}

// Close is close the Runtime.
func (r *Runtime) Close(ctx context.Context) error {
	if err := r.reconciler.Close(); err != nil {
		return err
	}
	if _, err := r.loader.UnloadMany(ctx, nil); err != nil {
		return err
	}
	return r.table.Close()
}
