package store

import "github.com/siyul-park/uniflow/pkg/types"

type Builder struct {
	field string
}

// Where creates a new Builder with a field
func Where(field string) *Builder {
	return &Builder{field: field}
}

func Set(v types.Map) types.Map   { return types.NewMap(types.NewString("$set"), v) }
func Unset(v types.Map) types.Map { return types.NewMap(types.NewString("$unset"), v) }

func And(filters ...types.Map) types.Map { return combine("$and", filters...) }
func Or(filters ...types.Map) types.Map  { return combine("$or", filters...) }

func (b *Builder) Equal(v types.Value) types.Map              { return b.condition("$eq", v) }
func (b *Builder) NotEqual(v types.Value) types.Map           { return b.condition("$ne", v) }
func (b *Builder) GreaterThan(v types.Value) types.Map        { return b.condition("$gt", v) }
func (b *Builder) GreaterThanOrEqual(v types.Value) types.Map { return b.condition("$gte", v) }
func (b *Builder) LessThan(v types.Value) types.Map           { return b.condition("$lt", v) }
func (b *Builder) LessThanOrEqual(v types.Value) types.Map    { return b.condition("$lte", v) }

func (b *Builder) condition(op string, v types.Value) types.Map {
	return types.NewMap(types.NewString(b.field), types.NewMap(types.NewString(op), v))
}

func combine(op string, filters ...types.Map) types.Map {
	children := make([]types.Value, 0, len(filters))
	for _, filter := range filters {
		if filter != nil {
			children = append(children, filter)
		}
	}
	if len(children) == 0 {
		return nil
	}
	if len(children) == 1 {
		return children[0].(types.Map)
	}
	return types.NewMap(types.NewString(op), types.NewSlice(children...))
}
