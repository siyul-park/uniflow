package loader

import (
	"context"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// ReconcilerConfig holds the configuration settings for the Reconciler.
type ReconcilerConfig struct {
	Namespace string        // Namespace associated with the Reconciler
	Storage   *spec.Storage // Storage used for watching changes to spec.Spec
	Loader    *Loader       // Loader to load spec.Spec into the symbol.Table
}

// Reconciler tracks changes to spec.Spec and keeps the symbol.Table up to date.
type Reconciler struct {
	namespace string
	storage   *spec.Storage
	loader    *Loader
	stream    *spec.Stream
	mu        sync.RWMutex
}

// NewReconciler creates a new Reconciler instance with the given configuration.
func NewReconciler(config ReconcilerConfig) *Reconciler {
	return &Reconciler{
		namespace: config.Namespace,
		storage:   config.Storage,
		loader:    config.Loader,
	}
}

// Watch starts watching for changes to spec.Spec.
func (r *Reconciler) Watch(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stream != nil {
		return nil
	}

	var filter *spec.Filter
	if r.namespace != "" {
		filter = spec.Where[string](spec.KeyNamespace).EQ(r.namespace)
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

// Reconcile reflects changes to spec.Spec in the symbol.Table.
func (r *Reconciler) Reconcile(ctx context.Context) error {
	stream := func() *spec.Stream {
		r.mu.RLock()
		defer r.mu.RUnlock()

		return r.stream
	}()
	if stream == nil {
		return nil
	}

	var next []uuid.UUID
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-stream.Next():
			if !ok {
				return nil
			}

			next = append(next, event.NodeID)

			for i := len(next) - 1; i >= 0; i-- {
				id := next[i]
				if _, err := r.loader.LoadOne(ctx, id); err == nil {
					next = append(next[:i], next[i+1:]...)
				}
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
