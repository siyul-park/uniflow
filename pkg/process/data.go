package process

import "sync"

// Data is a concurrent map-like data structure.
type Data struct {
	outer *Data
	data  map[string]any
	mu    sync.RWMutex
}

func newData() *Data {
	return &Data{
		data: make(map[string]any),
	}
}

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

func (d *Data) Store(key string, val any) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.data[key] = val
}

func (d *Data) Delete(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.data[key]; ok {
		delete(d.data, key)
		return true
	}
	return false
}

func (d *Data) Fork() *Data {
	return &Data{
		data:  make(map[string]any),
		outer: d,
	}
}

func (d *Data) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.data = make(map[string]any)
	d.outer = nil
}
