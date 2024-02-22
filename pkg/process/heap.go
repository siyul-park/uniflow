package process

import "sync"

type Heap struct {
	data map[string]any
	mu   sync.RWMutex
}

func newHeap() *Heap {
	return &Heap{
		data: make(map[string]any),
	}
}

func (h *Heap) Load(key string) any {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.data[key]
}

func (h *Heap) Store(key string, val any) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.data[key]; ok {
		return false
	}
	h.data[key] = val
	return true
}

func (h *Heap) Delete(key string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.data[key]; !ok {
		return false
	}
	delete(h.data, key)
	return true
}

func (h *Heap) LoadAndDelete(key string) any {
	h.mu.Lock()
	defer h.mu.Unlock()

	val := h.data[key]
	delete(h.data, key)
	return val
}

func (h *Heap) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.data = make(map[string]any)
}
