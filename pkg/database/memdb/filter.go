package memdb

import (
	"strconv"
	"strings"

	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/types"
)

func parseFilter(filter *database.Filter) func(types.Map) bool {
	if filter == nil {
		return nil
	}

	switch filter.OP {
	case database.EQ:
		return func(m types.Map) bool {
			if o, ok := types.Pick[types.Object](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return types.Equal(o, filter.Value)
			}
		}
	case database.NE:
		return func(m types.Map) bool {
			if o, ok := types.Pick[types.Object](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return !types.Equal(o, filter.Value)
			}
		}
	case database.LT:
		return func(m types.Map) bool {
			if o, ok := types.Pick[types.Object](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return types.Compare(o, filter.Value) < 0
			}
		}
	case database.LTE:
		return func(m types.Map) bool {
			if o, ok := types.Pick[types.Object](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return types.Compare(o, filter.Value) <= 0
			}
		}
	case database.GT:
		return func(m types.Map) bool {
			if o, ok := types.Pick[types.Object](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return types.Compare(o, filter.Value) > 0
			}
		}
	case database.GTE:
		return func(m types.Map) bool {
			if o, ok := types.Pick[types.Object](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return types.Compare(o, filter.Value) >= 0
			}
		}
	case database.IN:
		return func(m types.Map) bool {
			if o, ok := types.Pick[types.Object](m, parsePath(filter.Key)...); !ok {
				return false
			} else if o == nil {
				return false
			} else if v, ok := filter.Value.(types.Slice); !ok {
				return false
			} else {
				for i := 0; i < v.Len(); i++ {
					if types.Equal(o, v.Get(i)) {
						return true
					}
				}
				return false
			}
		}
	case database.NIN:
		return func(m types.Map) bool {
			if o, ok := types.Pick[types.Object](m, parsePath(filter.Key)...); !ok {
				return true
			} else if o == nil {
				return true
			} else if v, ok := filter.Value.(types.Slice); !ok {
				return false
			} else {
				for i := 0; i < v.Len(); i++ {
					if types.Equal(o, v.Get(i)) {
						return false
					}
				}
				return true
			}
		}
	case database.NULL:
		return func(m types.Map) bool {
			if v, ok := types.Pick[types.Object](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return v == nil
			}
		}
	case database.NNULL:
		return func(m types.Map) bool {
			if v, ok := types.Pick[types.Object](m, parsePath(filter.Key)...); !ok {
				return false
			} else {
				return v != nil
			}
		}
	case database.AND:
		parsed := make([]func(types.Map) bool, len(filter.Children))
		for i, child := range filter.Children {
			parsed[i] = parseFilter(child)
		}
		return func(m types.Map) bool {
			for _, p := range parsed {
				if !p(m) {
					return false
				}
			}
			return true
		}
	case database.OR:
		parsed := make([]func(types.Map) bool, len(filter.Children))
		for i, child := range filter.Children {
			parsed[i] = parseFilter(child)
		}
		return func(m types.Map) bool {
			for _, p := range parsed {
				if p(m) {
					return true
				}
			}
			return false
		}
	}

	return func(_ types.Map) bool {
		return false
	}
}

func extractIDByFilter(filter *database.Filter) types.Object {
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
		var id types.Object
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

func parsePath(key string) []any {
	tokens := strings.Split(key, ".")
	paths := make([]any, 0, len(tokens))

	for _, token := range tokens {
		if strings.Contains(token, "[") && strings.Contains(token, "]") {
			start := strings.Index(token, "[")
			end := strings.Index(token, "]")

			key1 := token[:start]
			key2 := token[start+1 : end]

			if index, err := strconv.Atoi(key2); err == nil {
				paths = append(paths, key1, index)
			} else if strings.HasPrefix(key2, "\"") && strings.HasSuffix(key2, "\"") {
				key2 = key2[1 : len(key2)-1]
				paths = append(paths, key1, key2)
			} else {
				paths = append(paths, key1, key2)
			}
		} else {
			paths = append(paths, token)
		}
	}

	return paths
}
