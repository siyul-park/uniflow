package chart

import (
	"context"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/value"
)

// LoaderConfig holds configuration for the Loader.
type LoaderConfig struct {
	Table      *Table      // Lookup table for storing loaded symbols
	ChartStore Store       // ChartStore to retrieve charts from
	ValueStore value.Store // ValueStore to retrieve values from
}

// Loader synchronizes with spec.Store to load spec.Spec into the Table.
type Loader struct {
	table      *Table
	chartStore Store
	valueStore value.Store
}

// NewLoader creates a new Loader instance with the provided configuration.
func NewLoader(config LoaderConfig) *Loader {
	return &Loader{
		table:      config.Table,
		chartStore: config.ChartStore,
		valueStore: config.ValueStore,
	}
}

// Load loads charts and binds them with values, then inserts them into the table.
func (l *Loader) Load(ctx context.Context, charts ...*Chart) error {
	examples := charts

	charts, err := l.chartStore.Load(ctx, examples...)
	if err != nil {
		return err
	}

	var values []*value.Value
	for _, chrt := range charts {
		for _, vals := range chrt.GetEnv() {
			for _, val := range vals {
				if val.ID == uuid.Nil && val.Name == "" {
					continue
				}
				values = append(values, &value.Value{
					ID:        val.ID,
					Namespace: chrt.GetNamespace(),
					Name:      val.Name,
				})
			}
		}
	}

	if len(values) > 0 {
		if values, err = l.valueStore.Load(ctx, values...); err != nil {
			return err
		}
	}

	var errs []error
	loaded := make([]*Chart, 0, len(charts))
	for _, chrt := range charts {
		if err := chrt.Bind(values...); err != nil {
			errs = append(errs, err)
		} else if err := l.table.Insert(chrt); err != nil {
			errs = append(errs, err)
		} else {
			loaded = append(loaded, chrt)
		}
	}

	for _, id := range l.table.Keys() {
		chrt := l.table.Lookup(id)
		if chrt != nil && len(resource.Match(chrt, examples...)) > 0 {
			ok := false
			for _, c := range loaded {
				if c.GetID() == id {
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
