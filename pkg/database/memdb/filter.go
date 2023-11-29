package memdb

import (
	"regexp"
	"strings"

	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

var (
	numberSubPath = regexp.MustCompile(`\[([0-9]+)\]`)
)

func ParseFilter(filter *database.Filter) func(*primitive.Map) bool {
	if filter == nil {
		return func(_ *primitive.Map) bool {
			return true
		}
	}

	switch filter.OP {
	case database.EQ:
		return func(m *primitive.Map) bool {
			if o, ok := primitive.Pick[primitive.Object](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return primitive.Compare(o, filter.Value) == 0
			}
		}
	case database.NE:
		return func(m *primitive.Map) bool {
			if o, ok := primitive.Pick[primitive.Object](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return primitive.Compare(o, filter.Value) != 0
			}
		}
	case database.LT:
		return func(m *primitive.Map) bool {
			if o, ok := primitive.Pick[primitive.Object](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return primitive.Compare(o, filter.Value) < 0
			}
		}
	case database.LTE:
		return func(m *primitive.Map) bool {
			if o, ok := primitive.Pick[primitive.Object](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return primitive.Compare(o, filter.Value) <= 0
			}
		}
	case database.GT:
		return func(m *primitive.Map) bool {
			if o, ok := primitive.Pick[primitive.Object](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return primitive.Compare(o, filter.Value) > 0
			}
		}
	case database.GTE:
		return func(m *primitive.Map) bool {
			if o, ok := primitive.Pick[primitive.Object](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return primitive.Compare(o, filter.Value) >= 0
			}
		}
	case database.IN:
		return func(m *primitive.Map) bool {
			if o, ok := primitive.Pick[primitive.Object](m, parsePath(filter.Key)...); !ok {
				return false
			} else if o == nil {
				return false
			} else if v, ok := filter.Value.(*primitive.Slice); !ok {
				return false
			} else {
				for i := 0; i < v.Len(); i++ {
					if primitive.Compare(o, v.Get(i)) == 0 {
						return true
					}
				}
				return false
			}
		}
	case database.NIN:
		return func(m *primitive.Map) bool {
			if o, ok := primitive.Pick[primitive.Object](m, parsePath(filter.Key)...); !ok {
				return true
			} else if o == nil {
				return true
			} else if v, ok := filter.Value.(*primitive.Slice); !ok {
				return false
			} else {
				for i := 0; i < v.Len(); i++ {
					if primitive.Compare(o, v.Get(i)) != 0 {
						return true
					}
				}
				return false
			}
		}
	case database.NULL:
		return func(m *primitive.Map) bool {
			if v, ok := primitive.Pick[primitive.Object](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return v == nil
			}
		}
	case database.NNULL:
		return func(m *primitive.Map) bool {
			if v, ok := primitive.Pick[primitive.Object](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return v != nil
			}
		}
	case database.AND:
		parsed := make([]func(*primitive.Map) bool, len(filter.Children))
		for i, child := range filter.Children {
			parsed[i] = ParseFilter(child)
		}
		return func(m *primitive.Map) bool {
			for _, p := range parsed {
				if !p(m) {
					return false
				}
			}
			return true
		}
	case database.OR:
		parsed := make([]func(*primitive.Map) bool, len(filter.Children))
		for i, child := range filter.Children {
			parsed[i] = ParseFilter(child)
		}
		return func(m *primitive.Map) bool {
			for _, p := range parsed {
				if p(m) {
					return true
				}
			}
			return false
		}
	}

	return func(_ *primitive.Map) bool {
		return false
	}
}

func FilterToExample(filter *database.Filter) ([]*primitive.Map, bool) {
	if filter == nil {
		return nil, false
	}

	switch filter.OP {
	case database.EQ:
		return []*primitive.Map{primitive.NewMap(primitive.NewString(filter.Key), filter.Value)}, true
	case database.NE:
		return nil, false
	case database.LT:
		return nil, false
	case database.LTE:
		return nil, false
	case database.GT:
		return nil, false
	case database.GTE:
		return nil, false
	case database.IN:
		if children, ok := filter.Value.(*primitive.Slice); !ok {
			return nil, false
		} else {
			examples := make([]*primitive.Map, children.Len())
			for i := 0; i < children.Len(); i++ {
				examples[i] = primitive.NewMap(primitive.NewString(filter.Key), children.Get(i))
			}
			return examples, true
		}
	case database.NIN:
		return nil, false
	case database.NULL:
		return []*primitive.Map{primitive.NewMap(primitive.NewString(filter.Key), nil)}, true
	case database.NNULL:
		return nil, false
	case database.AND:
		example := primitive.NewMap()
		for _, child := range filter.Children {
			e, _ := FilterToExample(child)
			if len(e) == 0 {
			} else if len(e) == 1 {
				for _, k := range e[0].Keys() {
					v, _ := e[0].Get(k)

					if _, ok := example.Get(k); ok {
						return nil, true
					} else {
						example.Set(k, v)
					}
				}
			} else {
				return nil, false
			}
		}
		return []*primitive.Map{example}, true
	case database.OR:
		var examples []*primitive.Map
		for _, child := range filter.Children {
			if e, ok := FilterToExample(child); ok {
				examples = append(examples, e...)
			} else {
				return nil, false
			}
		}
		return examples, true
	}

	return nil, false
}

func parsePath(key string) []string {
	key = numberSubPath.ReplaceAllString(key, ".$1")
	return strings.Split(key, ".")
}
