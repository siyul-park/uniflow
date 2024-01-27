package loader

import (
	"context"
	"sync"

	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
)

// ReconcilerConfig holds the configuration for the Reconciler.
type ReconcilerConfig struct {
	Namespace string
	Storage   *storage.Storage
	Loader    *Loader
}

// Reconciler keeps the symbol.Table up to date by tracking changes to scheme.Spec.
type Reconciler struct {
	namespace string
	storage   *storage.Storage
	loader    *Loader
	stream    *storage.Stream
	mu        sync.RWMutex
}

// NewReconciler creates a new Reconciler with the given configuration.
func NewReconciler(config ReconcilerConfig) *Reconciler {
	return &Reconciler{
		namespace: config.Namespace,
		storage:   config.Storage,
		loader:    config.Loader,
	}
}

// Watch starts watching for changes to scheme.Spec.
func (r *Reconciler) Watch(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stream != nil {
		return nil
	}

	var filter *storage.Filter
	if r.namespace != "" {
		filter = storage.Where[string](scheme.KeyNamespace).EQ(r.namespace)
	}

	s, err := r.storage.Watch(ctx, filter)
	if err != nil {
		return err
	}

	go func() {
		<-s.Done()

		r.mu.Lock()
		defer r.mu.Unlock()

		if r.stream == s {
			r.stream = nil
		}
	}()

	r.stream = s
	return nil
}

// Reconcile reflects changes to scheme.Spec in the symbol.Table.
func (r *Reconciler) Reconcile(ctx context.Context) error {
	stream := func() *storage.Stream {
		r.mu.RLock()
		defer r.mu.RUnlock()

		return r.stream
	}()
	if stream == nil {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-stream.Next():
			if !ok {
				return nil
			}
			if _, err := r.loader.LoadOne(ctx, event.NodeID); err != nil {
				return err
			}
		}
	}
}

// Close stops the Reconciler and closes the associated stream.
func (r *Reconciler) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stream == nil {
		return nil
	}
	if err := r.stream.Close(); err != nil {
		return err
	}
	r.stream = nil

	return nil
}
