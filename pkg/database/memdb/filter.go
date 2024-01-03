package memdb

import (
	"regexp"
	"strings"

	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

var numberSubPath = regexp.MustCompile(`\[([0-9]+)\]`)

func parseFilter(filter *database.Filter) func(*primitive.Map) bool {
	if filter == nil {
		return func(_ *primitive.Map) bool {
			return true
		}
	}

	switch filter.OP {
	case database.EQ:
		return func(m *primitive.Map) bool {
			if o, ok := primitive.Pick[primitive.Value](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return primitive.Compare(o, filter.Value) == 0
			}
		}
	case database.NE:
		return func(m *primitive.Map) bool {
			if o, ok := primitive.Pick[primitive.Value](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return primitive.Compare(o, filter.Value) != 0
			}
		}
	case database.LT:
		return func(m *primitive.Map) bool {
			if o, ok := primitive.Pick[primitive.Value](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return primitive.Compare(o, filter.Value) < 0
			}
		}
	case database.LTE:
		return func(m *primitive.Map) bool {
			if o, ok := primitive.Pick[primitive.Value](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return primitive.Compare(o, filter.Value) <= 0
			}
		}
	case database.GT:
		return func(m *primitive.Map) bool {
			if o, ok := primitive.Pick[primitive.Value](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return primitive.Compare(o, filter.Value) > 0
			}
		}
	case database.GTE:
		return func(m *primitive.Map) bool {
			if o, ok := primitive.Pick[primitive.Value](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return primitive.Compare(o, filter.Value) >= 0
			}
		}
	case database.IN:
		return func(m *primitive.Map) bool {
			if o, ok := primitive.Pick[primitive.Value](m, parsePath(filter.Key)...); !ok {
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
			if o, ok := primitive.Pick[primitive.Value](m, parsePath(filter.Key)...); !ok {
				return true
			} else if o == nil {
				return true
			} else if v, ok := filter.Value.(*primitive.Slice); !ok {
				return false
			} else {
				for i := 0; i < v.Len(); i++ {
					if primitive.Compare(o, v.Get(i)) == 0 {
						return false
					}
				}
				return true
			}
		}
	case database.NULL:
		return func(m *primitive.Map) bool {
			if v, ok := primitive.Pick[primitive.Value](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return v == nil
			}
		}
	case database.NNULL:
		return func(m *primitive.Map) bool {
			if v, ok := primitive.Pick[primitive.Value](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return v != nil
			}
		}
	case database.AND:
		parsed := make([]func(*primitive.Map) bool, len(filter.Children))
		for i, child := range filter.Children {
			parsed[i] = parseFilter(child)
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
			parsed[i] = parseFilter(child)
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

func extractIDByFilter(filter *database.Filter) primitive.Value {
	if filter == nil {
		return nil
	}

	switch filter.OP {
	case database.EQ:
		if filter.Key == keyID.String() {
			return filter.Value
		}
		return nil
	case database.AND:
		var id primitive.Value
		for _, child := range filter.Children {
			if childID := extractIDByFilter(child); childID != nil {
				if id != nil {
					return nil
				}
				id = childID
			}
		}
		return id
	default:
		return nil
	}
}

func parsePath(key string) []string {
	key = numberSubPath.ReplaceAllString(key, ".$1")
	return strings.Split(key, ".")
}
