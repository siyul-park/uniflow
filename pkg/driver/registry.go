package driver

import (
	"sync"

	"github.com/pkg/errors"
)

// Registry manages a collection of drivers.
type Registry struct {
	drivers map[string]Driver
	mu      sync.RWMutex
}

var (
	ErrAlreadyRegistered = errors.New("driver already registered")
	ErrNotRegistered     = errors.New("driver not registered")
)

// NewRegistry creates and returns a new Registry instance.
func NewRegistry() *Registry {
	return &Registry{drivers: make(map[string]Driver)}
}

// Register adds a driver to the registry. Returns an error if the driver already exists.
func (r *Registry) Register(name string, driver Driver) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.drivers[name]; ok {
		return errors.WithStack(ErrAlreadyRegistered)
	}
	r.drivers[name] = driver
	return nil
}

// Lookup retrieves a driver by its name. Returns an error if the driver is not found.
func (r *Registry) Lookup(name string) (Driver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	drv, ok := r.drivers[name]
	if !ok {
		return nil, errors.WithStack(ErrNotRegistered)
	}
	return drv, nil
}

// Close shuts down all registered drivers and clears the registry.
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, drv := range r.drivers {
		if err := drv.Close(); err != nil {
			return err
		}
	}
	r.drivers = make(map[string]Driver)
	return nil
}
