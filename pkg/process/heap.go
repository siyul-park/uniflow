package process

import "sync"

// Heap is a concurrent map-like data structure.
type Heap struct {
	data map[string]any
	mu   sync.RWMutex
}

func newHeap() *Heap {
	return &Heap{
		data: make(map[string]any),
	}
}

// Load retrieves the value associated with the given key from the heap.
func (h *Heap) Load(key string) any {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.data[key]
}

// Store stores the given value with the associated key in the heap.
func (h *Heap) Store(key string, val any) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.data[key]; ok {
		return false
	}
	h.data[key] = val
	return true
}

// Delete removes the value associated with the given key from the heap.
func (h *Heap) Delete(key string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.data[key]; !ok {
		return false
	}
	delete(h.data, key)
	return true
}

// LoadAndDelete retrieves and removes the value associated with the given key from the heap.
func (h *Heap) LoadAndDelete(key string) any {
	h.mu.Lock()
	defer h.mu.Unlock()

	val := h.data[key]
	delete(h.data, key)
	return val
}

// Close clears the heap by removing all key-value pairs.
func (h *Heap) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.data = make(map[string]any)
}
