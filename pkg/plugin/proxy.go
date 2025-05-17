package plugin

import (
	"context"
	"reflect"
	"strings"
)

type proxy struct {
	receiver reflect.Value
	methods  map[string]reflect.Value
}

var _ Plugin = (*proxy)(nil)

func (p *proxy) Name() string {
	m, ok := p.methods["Name"]
	if !ok || !m.IsValid() {
		return ""
	}

	t := m.Type()
	if t.NumIn() != 1 || t.NumOut() != 1 || t.Out(0).Kind() != reflect.String {
		return ""
	}

	ret := m.Call([]reflect.Value{p.receiver})
	return ret[0].Interface().(string)
}

func (p *proxy) Version() string {
	m, ok := p.methods["Version"]
	if !ok || !m.IsValid() {
		return ""
	}

	t := m.Type()
	if t.NumIn() != 1 || t.NumOut() != 1 || t.Out(0).Kind() != reflect.String {
		return ""
	}

	ret := m.Call([]reflect.Value{p.receiver})
	return ret[0].Interface().(string)
}

func (p *proxy) SetXXX(dep any) error {
	dv := reflect.ValueOf(dep)
	dt := dv.Type()

	for n, m := range p.methods {
		if !strings.HasPrefix(n, "Set") {
			continue
		}

		mv := m
		mt := mv.Type()

		if mt.NumIn() == 2 && dt.AssignableTo(mt.In(1)) {
			ret := mv.Call([]reflect.Value{p.receiver, dv})
			var err error
			if len(ret) > 0 {
				err, _ = ret[len(ret)-1].Interface().(error)
			}
			return err
		}
	}
	return nil
}

func (p *proxy) Load(ctx context.Context) error {
	m, ok := p.methods["Load"]
	if !ok || !m.IsValid() {
		return nil
	}

	t := m.Type()
	a0 := reflect.TypeOf(ctx)
	r0 := reflect.TypeOf((*error)(nil)).Elem()

	if t.NumIn() != 2 || !a0.AssignableTo(t.In(1)) || t.NumOut() != 1 || !t.Out(0).AssignableTo(r0) {
		return nil
	}

	ret := m.Call([]reflect.Value{p.receiver, reflect.ValueOf(ctx)})
	v0, _ := ret[0].Interface().(error)
	return v0
}

func (p *proxy) Unload(ctx context.Context) error {
	m, ok := p.methods["Unload"]
	if !ok || !m.IsValid() {
		return nil
	}

	t := m.Type()
	a0 := reflect.TypeOf(ctx)
	r0 := reflect.TypeOf((*error)(nil)).Elem()

	if t.NumIn() != 2 || !a0.AssignableTo(t.In(1)) || t.NumOut() != 1 || !t.Out(0).AssignableTo(r0) {
		return nil
	}

	ret := m.Call([]reflect.Value{p.receiver, reflect.ValueOf(ctx)})
	v0, _ := ret[0].Interface().(error)
	return v0
}
