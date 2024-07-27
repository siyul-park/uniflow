package symbol

import (
	"context"
	"reflect"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// LoaderConfig holds configuration for the Loader.
type LoaderConfig struct {
	Table       *Table         // Symbol table for storing loaded symbols
	Scheme      *scheme.Scheme // Scheme for decoding and compiling specs
	SpecStore   spec.Store     // SpecStore to retrieve specs from
	SecretStore secret.Store   // SecretStore to retrieve secrets from
}

// Loader synchronizes with spec.Store to load spec.Spec into the Table.
type Loader struct {
	table        *Table
	scheme       *scheme.Scheme
	specStore    spec.Store
	secretStore  secret.Store
	specStream   spec.Stream
	secretStream secret.Stream
	mu           sync.RWMutex
}

// NewLoader creates a new Loader instance with the provided configuration.
func NewLoader(config LoaderConfig) *Loader {
	return &Loader{
		table:       config.Table,
		scheme:      config.Scheme,
		specStore:   config.SpecStore,
		secretStore: config.SecretStore,
	}
}

// Load loads a spec.Spec by ID and its linked specs into the symbol table.
func (l *Loader) Load(ctx context.Context, specs ...spec.Spec) ([]*Symbol, error) {
	var symbols []*Symbol

	queue := specs
	for len(queue) > 0 {
		examples := queue
		queue = nil

		specs, err := l.specStore.Load(ctx, examples...)
		if err != nil {
			return nil, err
		}

		for _, spc := range specs {
			var secrets []*secret.Secret
			for _, values := range spc.GetEnv() {
				for _, value := range values {
					if value.ID == uuid.Nil && value.Name == "" {
						continue
					}
					secrets = append(secrets, &secret.Secret{
						ID:        value.ID,
						Namespace: spc.GetNamespace(),
						Name:      value.Name,
					})
				}
			}

			secrets, err := l.secretStore.Load(ctx, secrets...)
			if err != nil {
				return nil, err
			}

			bind, err := l.scheme.Bind(spc, secrets...)
			if err != nil {
				return nil, err
			}
			if bind == nil {
				if _, err := l.table.Free(spc.GetID()); err != nil {
					return nil, err
				}
				continue
			}

			decode, err := l.scheme.Decode(bind)
			if err != nil {
				return nil, err
			}

			sym, ok := l.table.Lookup(decode.GetID())
			if !ok || !reflect.DeepEqual(sym.Spec, decode) {
				n, err := l.scheme.Compile(decode)
				if err != nil {
					return nil, err
				}

				sym = &Symbol{Spec: decode, Node: n}
				if err := l.table.Insert(sym); err != nil {
					return nil, err
				}

				for _, ports := range sym.Ports() {
					for _, port := range ports {
						queue = append(queue, &spec.Meta{
							ID:        port.ID,
							Namespace: spc.GetNamespace(),
							Name:      port.Name,
						})
					}
				}
			}

			symbols = append(symbols, sym)
		}

		for _, example := range examples {
			if len(spec.Match(example, specs...)) == 0 {
				for _, id := range l.table.Keys() {
					sym, ok := l.table.Lookup(id)
					if ok && len(spec.Match(sym.Spec, example)) > 0 {
						if _, err := l.table.Free(sym.ID()); err != nil {
							return nil, err
						}
					}
				}
			}
		}
	}

	return symbols, nil
}

// Watch starts watching for changes to spec.Spec.
func (l *Loader) Watch(ctx context.Context, specs ...spec.Spec) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.specStream == nil {
		specStream, err := l.specStore.Watch(ctx, specs...)
		if err != nil {
			return err
		}

		go func() {
			<-specStream.Done()

			l.mu.Lock()
			defer l.mu.Unlock()

			if l.specStream == specStream {
				l.specStream = nil
			}
		}()

		l.specStream = specStream
	}

	if l.secretStream == nil {
		secrets := make([]*secret.Secret, len(specs))
		for _, spec := range specs {
			secrets = append(secrets, &secret.Secret{
				Namespace: spec.GetNamespace(),
			})
		}

		secretStream, err := l.secretStore.Watch(ctx, secrets...)
		if err != nil {
			return err
		}

		go func() {
			<-secretStream.Done()

			l.mu.Lock()
			defer l.mu.Unlock()

			if l.secretStream == secretStream {
				l.secretStream = nil
			}
		}()

		l.secretStream = secretStream
	}

	return nil
}

// Reconcile syncs changes to spec.Spec in the symbol table.
func (l *Loader) Reconcile(ctx context.Context) error {
	l.mu.RLock()

	specStream := l.specStream
	secretStream := l.secretStream

	l.mu.RUnlock()

	if specStream == nil || secretStream == nil {
		return nil
	}

	unloaded := map[uuid.UUID]spec.Spec{}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-secretStream.Next():
			if !ok {
				return nil
			}

			source := &secret.Secret{ID: event.ID}
			secrets, err := l.secretStore.Load(ctx, source)
			if err != nil {
				return err
			}
			if len(secrets) == 0 {
				secrets = append(secrets, source)
			}

			var examples []spec.Spec
			for _, id := range l.table.Keys() {
				sym, ok := l.table.Lookup(id)
				if !ok {
					continue
				}
				if l.scheme.IsBound(sym.Spec, secrets...) {
					examples = append(examples, sym.Spec)
				}
			}
			for _, example := range unloaded {
				if l.scheme.IsBound(example, secrets...) {
					examples = append(examples, example)
				}
			}

			symbols, err := l.Load(ctx, examples...)
			if err != nil {
				return err
			}

			for _, sym := range symbols {
				delete(unloaded, sym.ID())
			}
		case event, ok := <-specStream.Next():
			if !ok {
				return nil
			}

			specs, err := l.specStore.Load(ctx, &spec.Meta{ID: event.ID})
			if err != nil {
				return err
			}

			for _, spec := range specs {
				unloaded[spec.GetID()] = spec
			}

			var examples []spec.Spec
			for _, example := range unloaded {
				examples = append(examples, &spec.Meta{ID: example.GetID()})
			}

			symbols, err := l.Load(ctx, examples...)
			if err != nil {
				return err
			}

			for _, sym := range symbols {
				delete(unloaded, sym.ID())
			}
		}
	}
}

// Close stops the loader and closes the associated stream.
func (l *Loader) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.specStream != nil {
		if err := l.specStream.Close(); err != nil {
			return err
		}
		l.specStream = nil
	}

	if l.secretStream != nil {
		if err := l.secretStream.Close(); err != nil {
			return err
		}
		l.secretStream = nil
	}

	return nil
}
