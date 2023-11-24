package util

func Ptr[T any](value T) *T {
	return &value
}

func UnPtr[T any](value *T) T {
	if value != nil {
		return *value
	}
	var zero T
	return zero
}

func PtrTo[S any, T any](value *S, convert func(S) T) *T {
	if value == nil {
		return nil
	}
	return Ptr(convert(UnPtr(value)))
}
