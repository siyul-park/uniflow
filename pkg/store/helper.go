package store

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/types"
)

func match(doc, filter types.Value) (bool, error) {
	f, ok := filter.(types.Map)
	if !ok {
		return types.Equal(doc, filter), nil
	}

	for k, value := range f.Range() {
		key, ok := k.(types.String)
		if !ok {
			return false, errors.WithMessagef(ErrUnsupportedType, "key: %v", k.Interface())
		}

		if !strings.HasPrefix(key.String(), "$") {
			d, ok := doc.(types.Map)
			if !ok {
				return false, errors.WithMessagef(ErrUnsupportedType, "doc: %v", doc.Interface())
			}

			ok, err := match(d.Get(key), value)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
			continue
		}

		switch key.String() {
		case "$exists":
			if reflect.ValueOf(value).IsZero() {
				return value == nil, nil
			}
			return value != nil, nil
		case "$eq":
			if !types.Equal(doc, value) {
				return false, nil
			}
		case "$ne":
			if types.Equal(doc, value) {
				return false, nil
			}
		case "$gt":
			if types.Compare(doc, value) <= 0 {
				return false, nil
			}
		case "$lt":
			if types.Compare(doc, value) >= 0 {
				return false, nil
			}
		case "$gte":
			if types.Compare(doc, value) < 0 {
				return false, nil
			}
		case "$lte":
			if types.Compare(doc, value) > 0 {
				return false, nil
			}
		case "$and":
			vals, ok := value.(types.Slice)
			if !ok {
				return false, errors.WithMessagef(ErrUnsupportedType, "value: %v", value.Interface())
			}
			for _, sub := range vals.Range() {
				match, err := match(doc, sub)
				if err != nil {
					return false, err
				}
				if !match {
					return false, nil
				}
			}
		case "$or":
			vals, ok := value.(types.Slice)
			if !ok {
				return false, errors.WithMessagef(ErrUnsupportedType, "value: %v", value.Interface())
			}
			for _, sub := range vals.Range() {
				match, err := match(doc, sub)
				if err != nil {
					return false, err
				}
				if match {
					return true, nil
				}
			}
		default:
			return false, errors.WithMessagef(ErrUnsupportedOperation, "operation: %v", key.String())
		}
	}
	return true, nil
}

func patch(doc, update types.Map) (types.Map, error) {
	doc = doc.Mutable()
	for k, value := range update.Range() {
		key, ok := k.(types.String)
		if !ok {
			return nil, errors.WithMessagef(ErrUnsupportedType, "key: %v", k.Interface())
		}

		switch key.String() {
		case "$set":
			val, ok := value.(types.Map)
			if !ok {
				return nil, errors.WithMessagef(ErrUnsupportedType, "value: %v", value.Interface())
			}
			for k, v := range val.Range() {
				doc.Set(k, v)
			}
		case "$unset":
			val, ok := value.(types.Map)
			if !ok {
				return nil, errors.WithMessagef(ErrUnsupportedType, "value: %v", value.Interface())
			}
			for k := range val.Range() {
				doc.Delete(k)
			}
		default:
			return nil, errors.WithMessagef(ErrUnsupportedOperation, "operation: %v", key.String())
		}
	}
	return doc.Immutable(), nil
}

func apply(filter types.Value) (types.Value, error) {
	f, ok := filter.(types.Map)
	if !ok {
		return filter, nil
	}

	doc := types.NewMap().Mutable()

	for k, value := range f.Range() {
		key, ok := k.(types.String)
		if !ok {
			continue
		}

		if !strings.HasPrefix(key.String(), "$") {
			child, err := apply(value)
			if err != nil {
				return nil, err
			}
			doc = doc.Set(key, child)
			continue
		}

		switch key.String() {
		case "$eq":
			return value, nil
		case "$and", "$or":
			vals, ok := value.(types.Slice)
			if !ok {
				return nil, errors.WithMessagef(ErrUnsupportedType, "value: %v", value.Interface())
			}
			for _, sub := range vals.Range() {
				child, err := types.Cast[types.Map](apply(sub))
				if err != nil {
					return nil, err
				}

				for key, val := range child.Range() {
					if doc.Has(key) {
						return nil, errors.WithMessagef(ErrUnsupportedOperation, "key: %v", key.Interface())
					}
					doc = doc.Set(key, val)
				}
			}
		default:
			return nil, errors.WithMessagef(ErrUnsupportedOperation, "operation: %v", key.String())
		}
	}

	return doc.Immutable(), nil
}
