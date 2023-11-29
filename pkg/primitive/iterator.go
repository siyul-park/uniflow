package primitive

import "reflect"

func getCommonType(values []any) reflect.Type {
	if len(values) == 0 {
		return typeAny
	}

	commonType := reflect.TypeOf(values[0])
	for _, value := range values {
		typ := reflect.TypeOf(value)
		if typ != commonType {
			return typeAny
		}
	}

	return commonType
}
