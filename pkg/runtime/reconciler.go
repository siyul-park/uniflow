package runtime

import (
	"context"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// ReconcilerConfig holds the configuration for the Reconciler.
type ReconcilerConfig struct {
	Scheme       *scheme.Scheme
	SpecStore    spec.Store
	SecretStore  secret.Store
	SymbolTable  *symbol.Table
	SymbolLoader *symbol.Loader
}

// Reconciler is responsible for reconciling resources and managing state.
type Reconciler struct {
	scheme       *scheme.Scheme
	specStore    spec.Store
	secretStore  secret.Store
	specStream   spec.Stream
	secretStream secret.Stream
	symbolTable  *symbol.Table
	symbolLoader *symbol.Loader
	mu           sync.RWMutex
}

// NewReconiler creates a new instance of Reconciler with the provided configuration.
func NewReconiler(config ReconcilerConfig) *Reconciler {
	return &Reconciler{
		scheme:       config.Scheme,
		specStore:    config.SpecStore,
		secretStore:  config.SecretStore,
		symbolTable:  config.SymbolTable,
		symbolLoader: config.SymbolLoader,
	}
}

// Watch starts watching the specified resources for updates.
func (r *Reconciler) Watch(ctx context.Context, resources ...resource.Resource) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.specStream == nil {
		specs := make([]spec.Spec, 0, len(resources))
		for _, rsc := range resources {
			specs = append(specs, &spec.Meta{
				Namespace: rsc.GetNamespace(),
			})
		}

		specStream, err := r.specStore.Watch(ctx, specs...)
		if err != nil {
			return err
		}

		r.specStream = specStream

		go func() {
			<-specStream.Done()

			r.mu.Lock()
			defer r.mu.Unlock()

			if r.specStream == specStream {
				r.specStream = nil
			}
		}()
	}

	if r.secretStream == nil {
		secrets := make([]*secret.Secret, 0, len(resources))
		for _, rsc := range resources {
			secrets = append(secrets, &secret.Secret{
				Namespace: rsc.GetNamespace(),
			})
		}

		secretStream, err := r.secretStore.Watch(ctx, secrets...)
		if err != nil {
			return err
		}

		r.secretStream = secretStream

		go func() {
			<-secretStream.Done()

			r.mu.Lock()
			defer r.mu.Unlock()

			if r.secretStream == secretStream {
				r.secretStream = nil
			}
		}()
	}

	return nil
}

// Reconcile processes updates from the specification and secret streams.
func (r *Reconciler) Reconcile(ctx context.Context) error {
	r.mu.RLock()

	specStream := r.specStream
	secretStream := r.secretStream

	r.mu.RUnlock()

	if specStream == nil || secretStream == nil {
		return nil
	}

	buffer := map[uuid.UUID]spec.Spec{}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-secretStream.Next():
			if !ok {
				return nil
			}

			example := &secret.Secret{ID: event.ID}
			secrets, err := r.secretStore.Load(ctx, example)
			if err != nil {
				return err
			}
			if len(secrets) == 0 {
				secrets = append(secrets, example)
			}

			var examples []spec.Spec
			for _, id := range r.symbolTable.Keys() {
				sb, ok := r.symbolTable.Lookup(id)
				if ok && r.scheme.IsBound(sb.Spec, secrets...) {
					examples = append(examples, sb.Spec)
				}
			}
			for _, rsc := range buffer {
				if r.scheme.IsBound(rsc, secrets...) {
					examples = append(examples, rsc)
				}
			}

			for _, example := range examples {
				if _, err := r.symbolLoader.Load(ctx, &spec.Meta{ID: example.GetID()}); err == nil {
					delete(buffer, example.GetID())
				} else {
					buffer[example.GetID()] = example
				}
			}
		case event, ok := <-specStream.Next():
			if !ok {
				return nil
			}

			example := &spec.Meta{ID: event.ID}
			specs, err := r.specStore.Load(ctx, example)
			if err != nil {
				return err
			}
			if len(specs) == 0 {
				specs = append(specs, example)
			}

			for _, spec := range specs {
				buffer[spec.GetID()] = spec
			}

			var examples []spec.Spec
			for _, example := range buffer {
				examples = append(examples, example)
			}

			for _, example := range examples {
				if _, err := r.symbolLoader.Load(ctx, &spec.Meta{ID: example.GetID()}); err == nil {
					delete(buffer, example.GetID())
				}
			}
		}
	}
}

// Close stops watching the resources and cleans up any resources held.
func (r *Reconciler) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.specStream != nil {
		if err := r.specStream.Close(); err != nil {
			return err
		}
		r.specStream = nil
	}

	if r.secretStream != nil {
		if err := r.secretStream.Close(); err != nil {
			return err
		}
		r.secretStream = nil
	}

	return nil
}
