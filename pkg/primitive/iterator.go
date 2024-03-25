package primitive

import "reflect"

func getCommonType(values []any) reflect.Type {
	if len(values) == 0 {
		return typeAny
	}

	commonType := safeTypeOf(values[0])
	for _, value := range values {
		typ := safeTypeOf(value)
		if typ != commonType {
			return typeAny
		}
	}
	return commonType
}

func safeTypeOf(value any) reflect.Type {
	typ := reflect.TypeOf(value)
	if typ == nil {
		typ = typeAny
	}
	return typ
}
