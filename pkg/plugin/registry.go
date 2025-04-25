package plugin

import (
	"context"
	"sync"
)

type Registry struct {
	proxies []*Proxy
	mu      sync.RWMutex
}

var _ Plugin = (*Registry)(nil)

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Register(p Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.proxies = append(r.proxies, NewProxy(p))
	return nil
}

func (r *Registry) Set(dependencies ...any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, p := range r.proxies {
		if err := p.Set(dependencies...); err != nil {
			return err
		}
	}
	return nil
}

func (r *Registry) Load(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := 0; i < len(r.proxies); i++ {
		if err := r.proxies[i].Load(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *Registry) Unload(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := len(r.proxies) - 1; i >= 0; i-- {
		if err := r.proxies[i].Unload(ctx); err != nil {
			return err
		}
	}
	return nil
}
