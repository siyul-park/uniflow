package loader

import (
	"context"
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
)

type (
	// ReconcilerConfig is a config for for the Reconciler.
	ReconcilerConfig struct {
		Remote *storage.Storage
		Loader *Loader
		Filter *storage.Filter
	}

	// Reconciler keeps up to date symbol.Table by tracking changes to the scheme.Spec.
	Reconciler struct {
		remote *storage.Storage
		loader *Loader
		filter *storage.Filter
		stream *storage.Stream
		done   chan struct{}
		mu     sync.Mutex
	}
)

// NewReconciler returns a new Reconciler.
func NewReconciler(config ReconcilerConfig) *Reconciler {
	remote := config.Remote
	loader := config.Loader
	filter := config.Filter

	return &Reconciler{
		remote: remote,
		loader: loader,
		filter: filter,
		done:   make(chan struct{}),
	}
}

// Watch starts to watch the changes.
func (r *Reconciler) Watch(ctx context.Context) error {
	_, err := r.watch(ctx)
	return err
}

// Reconcile starts to reflects the changes.
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

			if _, err := r.loader.LoadOne(ctx, storage.Where[ulid.ULID](scheme.KeyID).EQ(event.NodeID)); err != nil {
				return err
			}
		}
	}
}

// Close closes the Reconciler.
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
	s, err := r.remote.Watch(ctx, r.filter)
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
