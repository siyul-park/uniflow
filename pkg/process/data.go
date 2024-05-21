package process

import "sync"

// Data is a concurrent map-like data structure.
type Data struct {
	data map[string]any
	mu   sync.RWMutex
}

func newData() *Data {
	return &Data{
		data: make(map[string]any),
	}
}

// Load retrieves the value associated with the given key from the heap.
func (h *Data) Load(key string) any {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.data[key]
}

// Store stores the given value with the associated key in the heap.
func (h *Data) Store(key string, val any) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.data[key] = val
}

// Delete removes the value associated with the given key from the heap.
func (h *Data) Delete(key string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.data[key]; !ok {
		return false
	}
	delete(h.data, key)
	return true
}

// LoadAndDelete retrieves and removes the value associated with the given key from the heap.
func (h *Data) LoadAndDelete(key string) any {
	h.mu.Lock()
	defer h.mu.Unlock()

	val := h.data[key]
	delete(h.data, key)
	return val
}

// Close clears the heap by removing all key-value pairs.
func (h *Data) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.data = make(map[string]any)
}
