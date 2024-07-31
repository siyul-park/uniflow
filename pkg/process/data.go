package process

import "sync"

// Data represents a hierarchical data structure that supports concurrent access.
type Data struct {
	outer *Data
	data  map[string]any
	mu    sync.RWMutex
}

// newData creates a new Data instance.
func newData() *Data {
	return &Data{
		data: make(map[string]any),
	}
}

// LoadAndDelete retrieves and deletes the value for the given key.
// If not found, it checks the outer Data instance.
func (d *Data) LoadAndDelete(key string) any {
	d.mu.Lock()
	defer d.mu.Unlock()

	if val, ok := d.data[key]; ok {
		delete(d.data, key)
		return val
	}

	if d.outer == nil {
		return nil
	}
	return d.outer.LoadAndDelete(key)
}

// Load retrieves the value for the given key.
// If not found, it checks the outer Data instance.
func (d *Data) Load(key string) any {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if val, ok := d.data[key]; ok {
		return val
	}

	if d.outer == nil {
		return nil
	}
	return d.outer.Load(key)
}

// Store stores the value under the given key.
func (d *Data) Store(key string, val any) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.data[key] = val
}

// Delete removes the value for the given key.
// Returns true if the key existed and was deleted.
func (d *Data) Delete(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.data[key]; ok {
		delete(d.data, key)
		return true
	}
	return false
}

// Fork creates a new Data instance that inherits from the current instance.
func (d *Data) Fork() *Data {
	return &Data{
		data:  make(map[string]any),
		outer: d,
	}
}

// Close clears the data and removes the reference to the outer instance.
func (d *Data) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.data = make(map[string]any)
	d.outer = nil
}
