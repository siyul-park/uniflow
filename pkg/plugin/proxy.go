package plugin

import (
	"context"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// Proxy wraps a Plugin and injects dependencies via Inject* methods.
type Proxy struct {
	plugin Plugin
}

var ErrMissingDependency = errors.New("missing dependency")

var _ Plugin = (*Proxy)(nil)

// NewProxy creates a new Proxy for the given Plugin.
func NewProxy(plugin Plugin) *Proxy {
	return &Proxy{plugin: plugin}
}

// Inject calls Set* methods on the plugin that accept the given dependency.
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
			if ret := mv.Call([]reflect.Value{dv}); len(ret) > 0 {
				if err, ok := ret[0].Interface().(error); ok && err != nil {
					return false, err
				}
			}
			return true, nil
		}
	}
	return false, nil
}

// Load calls the plugin's Load method.
func (p *Proxy) Load(ctx context.Context) error {
	return p.plugin.Load(ctx)
}

// Unload calls the plugin's Unload method.
func (p *Proxy) Unload(ctx context.Context) error {
	return p.plugin.Unload(ctx)
}

// Unwrap returns the original plugin instance.
func (p *Proxy) Unwrap() Plugin {
	return p.plugin
}
