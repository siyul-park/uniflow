package plugin

import (
	"context"
	"reflect"
	"strings"
)

// Proxy wraps a Plugin and supports dependency injection.
type Proxy struct {
	plugin Plugin
}

var _ Plugin = (*Proxy)(nil)

// NewProxy returns a new Proxy for the given Plugin.
func NewProxy(plugin Plugin) *Proxy {
	return &Proxy{plugin: plugin}
}

// Inject injects a dependency via a matching Set* method.
func (p *Proxy) Inject(dependency any) (bool, error) {
	pv := reflect.ValueOf(p.plugin)
	pt := pv.Type()

	dv := reflect.ValueOf(dependency)
	dt := dv.Type()

	for i := 0; i < pt.NumMethod(); i++ {
		m := pt.Method(i)
		if !strings.HasPrefix(m.Name, "Set") {
			continue
		}

		mv := pv.Method(i)
		mt := mv.Type()

		if mt.NumIn() == 1 && dt.AssignableTo(mt.In(0)) {
			ret := mv.Call([]reflect.Value{dv})
			if len(ret) > 0 {
				if err, ok := ret[0].Interface().(error); ok && err != nil {
					return false, err
				}
			}
			return true, nil
		}
	}
	return false, nil
}

// Name returns the plugin name.
func (p *Proxy) Name() string {
	return p.plugin.Name()
}

// Version returns the plugin version.
func (p *Proxy) Version() string {
	return p.plugin.Version()
}

// Load loads the plugin.
func (p *Proxy) Load(ctx context.Context) error {
	return p.plugin.Load(ctx)
}

// Unload unloads the plugin.
func (p *Proxy) Unload(ctx context.Context) error {
	return p.plugin.Unload(ctx)
}

// Unwrap returns the original plugin.
func (p *Proxy) Unwrap() Plugin {
	return p.plugin
}
