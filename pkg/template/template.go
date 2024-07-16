package template

import (
	"reflect"
	"text/template"
)

// Template represents a parsed template with a root node.
type Template struct {
	name string
	root node
}

// New creates a new Template with the given name.
func New(name string) *Template {
	return &Template{
		name: name,
	}
}

// Parse parses the provided value into the template's root node.
func (t *Template) Parse(value any) (*Template, error) {
	root, err := t.parse(reflect.ValueOf(value))
	if err != nil {
		return nil, err
	}
	t.root = root
	return t, nil
}

// Execute applies the template to the provided data.
func (t *Template) Execute(data any) (any, error) {
	if t.root == nil {
		return data, nil
	}
	return t.root.execute(data)
}

// parse recursively parses a reflect.Value into a corresponding node.
func (t *Template) parse(val reflect.Value) (node, error) {
	switch val.Kind() {
	case reflect.String:
		tmpl, err := template.New(t.name).Parse(val.String())
		if err != nil {
			return nil, err
		}
		return &templateNode{typ: val.Type(), template: tmpl}, nil
	case reflect.Slice, reflect.Array:
		children := make([]node, val.Len())
		for i := 0; i < val.Len(); i++ {
			child, err := t.parse(reflect.ValueOf(val.Index(i).Interface()))
			if err != nil {
				return nil, err
			}
			children[i] = child
		}
		return &sliceNode{typ: val.Type(), children: children}, nil
	case reflect.Map:
		children := make(map[node]node)
		for _, key := range val.MapKeys() {
			k, err := t.parse(reflect.ValueOf(key.Interface()))
			if err != nil {
				return nil, err
			}
			v, err := t.parse(reflect.ValueOf(val.MapIndex(key).Interface()))
			if err != nil {
				return nil, err
			}
			children[k] = v
		}
		return &mapNode{typ: val.Type(), children: children}, nil
	default:
		return &valueNode{value: val.Interface()}, nil
	}
}
