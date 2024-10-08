package symbol

import (
	"context"
	"errors"
	"reflect"

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
	table       *Table
	scheme      *scheme.Scheme
	specStore   spec.Store
	secretStore secret.Store
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
func (l *Loader) Load(ctx context.Context, specs ...spec.Spec) error {
	examples := specs

	specs, err := l.specStore.Load(ctx, examples...)
	if err != nil {
		return err
	}

	var secrets []*secret.Secret
	for _, sp := range specs {
		for _, vals := range sp.GetEnv() {
			for _, val := range vals {
				if val.ID == uuid.Nil && val.Name == "" {
					continue
				}
				secrets = append(secrets, &secret.Secret{
					ID:        val.ID,
					Namespace: sp.GetNamespace(),
					Name:      val.Name,
				})
			}
		}
	}

	if len(secrets) > 0 {
		secrets, err = l.secretStore.Load(ctx, secrets...)
		if err != nil {
			return err
		}
	}

	var symbols []*Symbol
	var errs []error
	for _, sp := range specs {
		bind, err := spec.Bind(sp, secrets...)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		decode, err := l.scheme.Decode(bind)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		sb := l.table.Lookup(decode.GetID())
		if sb == nil || !reflect.DeepEqual(sb.Spec, decode) {
			n, err := l.scheme.Compile(decode)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			sb = &Symbol{Spec: decode, Node: n}
			if err := l.table.Insert(sb); err != nil {
				errs = append(errs, err)
				continue
			}
		}

		symbols = append(symbols, sb)
	}

	if len(errs) > 0 {
		for _, sb := range symbols {
			sb.Close()
		}
		symbols = nil
	}

	for _, id := range l.table.Keys() {
		sb := l.table.Lookup(id)
		if sb != nil && len(resource.Match(sb.Spec, examples...)) > 0 {
			ok := false
			for _, s := range symbols {
				if s.ID() == id {
					ok = true
					break
				}
			}
			if !ok {
				if _, err := l.table.Free(id); err != nil {
					return err
				}
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
