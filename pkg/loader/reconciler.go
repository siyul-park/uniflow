package loader

import (
	"context"
	"sync"

	"github.com/siyul-park/uniflow/pkg/storage"
)

// ReconcilerConfig holds the configuration for the Reconciler.
type ReconcilerConfig struct {
	Storage *storage.Storage // Storage is the storage used by the Reconciler.
	Loader  *Loader          // Loader is used to load scheme.Spec into the symbol.Table.
	Filter  *storage.Filter  // Filter is the filter for tracking changes to the scheme.Spec.
}

// Reconciler keeps the symbol.Table up to date by tracking changes to scheme.Spec.
type Reconciler struct {
	storage *storage.Storage
	loader  *Loader
	filter  *storage.Filter
	stream  *storage.Stream
	done    chan struct{}
	mu      sync.Mutex
}

// NewReconciler creates a new Reconciler with the given configuration.
func NewReconciler(config ReconcilerConfig) *Reconciler {
	return &Reconciler{
		storage: config.Storage,
		loader:  config.Loader,
		filter:  config.Filter,
		done:    make(chan struct{}),
	}
}

// Watch starts watching for changes to scheme.Spec.
func (r *Reconciler) Watch(ctx context.Context) error {
	_, err := r.watch(ctx)
	return err
}

// Reconcile reflects changes to scheme.Spec in the symbol.Table.
func (r *Reconciler) Reconcile(ctx context.Context) error {
	stream, err := r.watch(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-r.done:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case <-stream.Done():
			stream, err = r.watch(ctx)
			if err != nil {
				return err
			}
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

	select {
	case <-r.done:
		return nil
	default:
	}

	if r.stream == nil {
		return nil
	}
	if err := r.stream.Close(); err != nil {
		return err
	}
	r.stream = nil
	close(r.done)
	return nil
}

func (r *Reconciler) watch(ctx context.Context) (*storage.Stream, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stream != nil {
		return r.stream, nil
	}
	s, err := r.storage.Watch(ctx, r.filter)
	if err != nil {
		return nil, err
	}

	go func() {
		select {
		case <-s.Done():
			r.mu.Lock()
			defer r.mu.Unlock()

			r.stream = nil
		case <-r.done:
			return
		}
	}()

	r.stream = s
	return s, nil
}
