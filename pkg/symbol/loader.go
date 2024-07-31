package symbol

import (
	"context"
	"errors"
	"reflect"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/resource"
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
	examples := specs
	specs, err := l.specStore.Load(ctx, examples...)
	if err != nil {
		return nil, err
	}

	var secrets []*secret.Secret
	for _, spc := range specs {
		for _, vals := range spc.GetEnv() {
			for _, val := range vals {
				if val.ID == uuid.Nil && val.Name == "" {
					continue
				}
				secrets = append(secrets, &secret.Secret{
					ID:        val.ID,
					Namespace: spc.GetNamespace(),
					Name:      val.Name,
				})
			}
		}
	}

	if len(secrets) > 0 {
		secrets, err = l.secretStore.Load(ctx, secrets...)
		if err != nil {
			return nil, err
		}
	}

	var symbols []*Symbol
	var errs []error
	for _, spc := range specs {
		bind, err := l.scheme.Bind(spc, secrets...)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		decode, err := l.scheme.Decode(bind)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		sym, ok := l.table.Lookup(decode.GetID())
		if !ok || !reflect.DeepEqual(sym.Spec, decode) {
			n, err := l.scheme.Compile(decode)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			sym = &Symbol{Spec: decode, Node: n}
			if err := l.table.Insert(sym); err != nil {
				errs = append(errs, err)
				continue
			}
		}

		symbols = append(symbols, sym)
	}

	if len(errs) > 0 {
		symbols = nil
	}

	for _, id := range l.table.Keys() {
		sym, ok := l.table.Lookup(id)
		if ok && len(resource.Match(sym.Spec, examples...)) > 0 {
			var sym *Symbol
			for _, s := range symbols {
				if s.ID() == id {
					sym = s
					break
				}
			}
			if sym == nil {
				if _, err := l.table.Free(id); err != nil {
					return nil, err
				}
			}
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
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

		l.specStream = specStream

		go func() {
			<-specStream.Done()

			l.mu.Lock()
			defer l.mu.Unlock()

			if l.specStream == specStream {
				l.specStream = nil
			}
		}()
	}

	if l.secretStream == nil {
		secrets := make([]*secret.Secret, 0, len(specs))
		for _, spc := range specs {
			secrets = append(secrets, &secret.Secret{
				Namespace: spc.GetNamespace(),
			})
		}

		secretStream, err := l.secretStore.Watch(ctx, secrets...)
		if err != nil {
			return err
		}

		l.secretStream = secretStream

		go func() {
			<-secretStream.Done()

			l.mu.Lock()
			defer l.mu.Unlock()

			if l.secretStream == secretStream {
				l.secretStream = nil
			}
		}()
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
			secrets, err := l.secretStore.Load(ctx, example)
			if err != nil {
				return err
			}
			if len(secrets) == 0 {
				secrets = append(secrets, example)
			}

			var examples []spec.Spec
			for _, id := range l.table.Keys() {
				sym, ok := l.table.Lookup(id)
				if ok && l.scheme.IsBound(sym.Spec, secrets...) {
					examples = append(examples, sym.Spec)
				}
			}
			for _, spc := range buffer {
				if l.scheme.IsBound(spc, secrets...) {
					examples = append(examples, spc)
				}
			}

			for _, example := range examples {
				if _, err := l.Load(ctx, &spec.Meta{ID: example.GetID()}); err == nil {
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
			specs, err := l.specStore.Load(ctx, example)
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
				if _, err := l.Load(ctx, &spec.Meta{ID: example.GetID()}); err == nil {
					delete(buffer, example.GetID())
				}
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
