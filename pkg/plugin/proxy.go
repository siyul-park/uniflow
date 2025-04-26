package plugin

import (
	"context"
	"reflect"
	"strings"
)

// Proxy wraps a Plugin and injects dependencies via Inject* methods.
type Proxy struct {
	plugin Plugin
}

var _ Plugin = (*Proxy)(nil)

// NewProxy creates a new Proxy for the given Plugin.
func NewProxy(plugin Plugin) *Proxy {
	return &Proxy{plugin: plugin}
}

// Inject injects dependencies into the plugin by calling its Inject* methods.
func (p *Proxy) Inject(dependencies ...any) error {
	val := reflect.ValueOf(p.plugin)
	for i := 0; i < val.NumMethod(); i++ {
		typ := val.Type().Method(i)
		val := val.Method(i)

		if !strings.HasPrefix(typ.Name, "Set") {
			continue
		}

		var ins []reflect.Value
		for j := 0; j < val.Type().NumIn(); j++ {
			typ := val.Type().In(j)
			for _, dep := range dependencies {
				if reflect.TypeOf(dep).AssignableTo(typ) {
					ins = append(ins, reflect.ValueOf(dep))
					break
				}
			}
		}

		outs := val.Call(ins)
		if len(outs) > 0 {
			if err, ok := outs[len(outs)-1].Interface().(error); ok && err != nil {
				return err
			}
		}
	}
	return nil
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
