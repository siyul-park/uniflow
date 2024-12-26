package types

// Get extracts a value from a nested structure using the provided paths.
func Get[T any](obj Value, paths ...any) (T, bool) {
	var val T
	cur := obj
	for _, path := range paths {
		p, err := Marshal(path)
		if err != nil {
			return val, false
		}

		switch p := p.(type) {
		case String:
			if v, ok := cur.(Map); ok {
				child := v.Get(p)
				if child == nil {
					return val, false
				}
				cur = child
			}
		case Integer:
			if v, ok := cur.(Slice); ok {
				if int(p.Int()) >= v.Len() {
					return val, false
				}
				cur = v.Get(int(p.Int()))
			}
		default:
			return val, false
		}
	}

	if cur == nil {
		return val, false
	}
	if v, ok := cur.(T); ok {
		return v, true
	}
	return val, Unmarshal(cur, &val) == nil
}
