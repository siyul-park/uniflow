package symbol

import (
	"context"
	"reflect"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/template"
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

	nexts := specs
	for len(nexts) > 0 {
		curr := nexts
		nexts = nil

		specs, err := l.specStore.Load(ctx, curr...)
		if err != nil {
			return nil, err
		}

		for _, s := range specs {
			var secrets []*secret.Secret
			for _, values := range s.GetEnv() {
				for _, value := range values {
					secrets = append(secrets, &secret.Secret{
						ID:        value.ID,
						Namespace: s.GetNamespace(),
						Name:      value.Name,
					})
				}
			}

			secrets, err := l.secretStore.Load(ctx, secrets...)
			if err != nil {
				return nil, err
			}

			if len(secrets) > 0 {
				ok := true
				for _, values := range s.GetEnv() {
					for i, value := range values {
						if value.ID == uuid.Nil && value.Name == "" {
							continue
						}

						secret := l.lookup(secrets, value)
						if secret == nil {
							ok = false
							break
						}

						if tmpl, err := template.New("").Parse(value.Value); err != nil {
							return nil, err
						} else if v, err := tmpl.Execute(secret.Data); err != nil {
							return nil, err
						} else {
							value.Value = v
							values[i] = value
						}
					}
					if !ok {
						break
					}
				}
				if !ok {
					continue
				}
			}

			decode, err := l.scheme.Decode(s)
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

				for _, locations := range sym.Ports() {
					for _, location := range locations {
						nexts = append(nexts, &spec.Meta{
							ID:        location.ID,
							Namespace: s.GetNamespace(),
							Name:      location.Name,
						})
					}
				}
			}

			symbols = append(symbols, sym)
		}

		for _, spec := range curr {
			exists := false
			for _, s := range specs {
				if spec.GetID() == s.GetID() || (spec.GetNamespace() == s.GetNamespace() && spec.GetName() == s.GetName()) {
					exists = true
					break
				}
				if exists {
					break
				}
			}

			if !exists {
				sym, ok := l.table.Lookup(spec.GetID())
				if ok {
					if _, err := l.table.Free(sym.ID()); err != nil {
						return nil, err
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

			secrets, err := l.secretStore.Load(ctx, &secret.Secret{ID: event.ID})
			if err != nil {
				return err
			}

			var examples []spec.Spec
			for _, id := range l.table.Keys() {
				sym, ok := l.table.Lookup(id)
				if !ok {
					continue
				}
				examples = append(examples, sym.Spec)
			}
			for _, example := range unloaded {
				examples = append(examples, example)
			}

			for i, example := range examples {
				ok := false
				for _, values := range example.GetEnv() {
					for _, value := range values {
						if secret := l.lookup(secrets, value); secret != nil {
							ok = true
							break
						}
					}
				}
				if !ok {
					examples = append(examples[:i], examples[i+1:]...)
					i--
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

func (l *Loader) lookup(secrets []*secret.Secret, secret spec.Secret) *secret.Secret {
	for _, s := range secrets {
		if s.GetID() == secret.ID || (s.GetName() != "" && s.GetName() == secret.Name) {
			return s
		}
	}
	return nil
}
