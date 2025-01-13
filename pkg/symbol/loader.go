package symbol

import (
	"context"
	"errors"
	"reflect"

	"github.com/iancoleman/strcase"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/value"
)

// LoaderConfig holds configuration for the Loader.
type LoaderConfig struct {
	Environment map[string]string // Environment holds the variables for the loader.
	Table       *Table            // Symbol table for storing loaded symbols
	Scheme      *scheme.Scheme    // Scheme for decoding and compiling specs
	SpecStore   spec.Store        // SpecStore to retrieve specs from
	ValueStore  value.Store       // ValueStore to retrieve values from
}

// Loader synchronizes with spec.Store to load spec.Spec into the Table.
type Loader struct {
	environment map[string]string
	table       *Table
	scheme      *scheme.Scheme
	specStore   spec.Store
	valueStore  value.Store
}

// NewLoader creates a new Loader instance with the provided configuration.
func NewLoader(config LoaderConfig) *Loader {
	env := map[string]string{}
	for k, v := range config.Environment {
		env[strcase.ToScreamingSnake(k)] = v
	}
	return &Loader{
		environment: env,
		table:       config.Table,
		scheme:      config.Scheme,
		specStore:   config.SpecStore,
		valueStore:  config.ValueStore,
	}
}

// Load loads a spec.Spec by ID and its linked specs into the symbol table.
func (l *Loader) Load(ctx context.Context, specs ...spec.Spec) error {
	examples := specs

	specs, err := l.specStore.Load(ctx, examples...)
	if err != nil {
		return err
	}

	var values []*value.Value
	for _, sp := range specs {
		for _, val := range sp.GetEnv() {
			if val.IsIdentified() {
				values = append(values, &value.Value{
					ID:        val.ID,
					Namespace: sp.GetNamespace(),
					Name:      val.Name,
				})
			}
		}
	}

	if len(values) > 0 {
		values, err = l.valueStore.Load(ctx, values...)
		if err != nil {
			return err
		}
	}

	if len(l.environment) > 0 {
		values = append(values, &value.Value{Data: l.environment})
	}

	var symbols []*Symbol
	var errs []error
	for _, sp := range specs {
		unstructured := &spec.Unstructured{}
		if err := spec.As(sp, unstructured); err != nil {
			errs = append(errs, err)
		} else if err := unstructured.Bind(values...); err != nil {
			errs = append(errs, err)
		} else if err := unstructured.Build(); err != nil {
			errs = append(errs, err)
		} else if decode, err := l.scheme.Decode(unstructured); err != nil {
			errs = append(errs, err)
		} else {
			sp = decode
		}

		sb := l.table.Lookup(sp.GetID())
		if sb == nil || !reflect.DeepEqual(sb.Spec, sp) {
			n, err := l.scheme.Compile(sp)
			if err != nil {
				errs = append(errs, err)
			}

			sb = &Symbol{Spec: sp, Node: n}
			if err := l.table.Insert(sb); err != nil {
				errs = append(errs, err)
			}
		}

		symbols = append(symbols, sb)
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
					errs = append(errs, err)
				}
			}
		}
	}
	return errors.Join(errs...)
}
