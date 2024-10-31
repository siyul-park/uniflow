package template

import (
	"bytes"
	"reflect"
	"text/template"
)

type node interface {
	execute(data any) (any, error)
}

type valueNode struct {
	value any
}

type templateNode struct {
	typ      reflect.Type
	template *template.Template
}

type sliceNode struct {
	typ      reflect.Type
	children []node
}

type mapNode struct {
	typ      reflect.Type
	children map[node]node
}

func (v *valueNode) execute(_ any) (any, error) {
	return v.value, nil
}

func (t *templateNode) execute(data any) (any, error) {
	var buf bytes.Buffer
	if err := t.template.Execute(&buf, data); err != nil {
		return nil, err
	}
	return reflect.ValueOf(buf.String()).Convert(t.typ).Interface(), nil
}

func (s *sliceNode) execute(data any) (any, error) {
	values := reflect.MakeSlice(s.typ, 0, len(s.children))
	for _, child := range s.children {
		value, err := child.execute(data)
		if err != nil {
			return nil, err
		}
		values = reflect.Append(values, reflect.ValueOf(value))
	}
	return values.Interface(), nil
}

func (m *mapNode) execute(data any) (any, error) {
	values := reflect.MakeMap(m.typ)
	for key, child := range m.children {
		keyRes, err := key.execute(data)
		if err != nil {
			return nil, err
		}
		value, err := child.execute(data)
		if err != nil {
			return nil, err
		}
		values.SetMapIndex(reflect.ValueOf(keyRes), reflect.ValueOf(value))
	}
	return values.Interface(), nil
}
