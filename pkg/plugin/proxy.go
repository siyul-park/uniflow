package plugin

import (
	"context"
	"reflect"
	"strings"
)

type Proxy struct {
	plugin Plugin
}

var _ Plugin = (*Proxy)(nil)

func NewProxy(plugin Plugin) *Proxy {
	return &Proxy{plugin: plugin}
}

func (p *Proxy) Set(dependencies ...any) error {
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
				if reflect.TypeOf(dep).ConvertibleTo(typ) {
					ins = append(ins, reflect.ValueOf(dep).Convert(typ))
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

func (p *Proxy) Load(ctx context.Context) error {
	return p.plugin.Load(ctx)
}

func (p *Proxy) Unload(ctx context.Context) error {
	return p.plugin.Unload(ctx)
}
