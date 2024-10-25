package process

// StoreHook represents an interface for storing values.
type StoreHook[T any] interface {
	// Store is called to store a value.
	Store(val T)
}

type storeHook[T any] struct {
	store func(val T)
}

// StoreHooks is a slice of StoreHook interfaces, processed in reverse order.
type StoreHooks[T any] []StoreHook[T]

var _ StoreHook[any] = (StoreHooks[any])(nil)
var _ StoreHook[any] = (*storeHook[any])(nil)

// StoreFunc creates a new StoreHook from the provided function.
func StoreFunc[T any](store func(val T)) StoreHook[T] {
	return &storeHook[T]{store: store}
}

// Store calls the Store method on each hook in reverse order.
func (h StoreHooks[T]) Store(val T) {
	for _, hook := range h {
		hook.Store(val)
	}
}

func (h *storeHook[T]) Store(val T) {
	h.store(val)
}
