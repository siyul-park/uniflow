package template

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// Values represents the environment of your application.
type Values struct {
	data map[string]any
	mu   sync.RWMutex
}

var typeAny = reflect.TypeOf((*any)(nil)).Elem()

// NewValues creates a new environment instance.
func NewValues() *Values {
	return &Values{
		data: make(map[string]any),
	}
}

// Set sets the value for the specified key in the environment.
func (vl *Values) Set(key string, val any) error {
	vl.mu.Lock()
	defer vl.mu.Unlock()

	tokens := strings.Split(key, ".")
	v := reflect.ValueOf(val)

	node := reflect.ValueOf(vl.data)
	for i, token := range tokens {
		t := reflect.ValueOf(token)

		if node.Kind() != reflect.Map || node.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("invalid map type for key %s", key)
		}

		if i == len(tokens)-1 {
			node.SetMapIndex(t, v)
			return nil
		}

		child := node.MapIndex(t)
		if !child.IsValid() {
			child = reflect.MakeMap(reflect.MapOf(reflect.TypeOf(""), typeAny))
			node.SetMapIndex(t, child)
		}
		node = reflect.ValueOf(child.Interface())
	}
	return nil
}

// Get retrieves the value for the specified key from the environment.
// If the key is not present, it returns a default value and a boolean indicating the key's existence.
func (vl *Values) Get(key string) (any, bool) {
	vl.mu.RLock()
	defer vl.mu.RUnlock()

	tokens := strings.Split(key, ".")

	node := reflect.ValueOf(vl.data)
	for _, token := range tokens {
		t := reflect.ValueOf(token)

		if node.Kind() != reflect.Map || node.Type().Key().Kind() != reflect.String {
			return nil, false
		}

		child := node.MapIndex(t)
		if !child.IsValid() {
			return nil, false
		}
		node = reflect.ValueOf(child.Interface())
	}

	return node.Interface(), true
}

// Delete removes the specified key from the environment.
func (vl *Values) Delete(key string) {
	vl.mu.Lock()
	defer vl.mu.Unlock()

	tokens := strings.Split(key, ".")
	vl.deleteRecursive(reflect.ValueOf(vl.data), tokens)
}

// GetData returns a read-only copy of the environment data.
func (vl *Values) GetData() map[string]any {
	vl.mu.RLock()
	defer vl.mu.RUnlock()

	result := make(map[string]any, len(vl.data))
	for k, v := range vl.data {
		result[k] = v
	}
	return result
}

func (vl *Values) deleteRecursive(node reflect.Value, tokens []string) bool {
	t := reflect.ValueOf(tokens[0])

	if node.Kind() != reflect.Map || node.Type().Key().Kind() != reflect.String {
		return false
	}

	if len(tokens) == 1 {
		child := node.MapIndex(t)
		if child.IsValid() {
			node.SetMapIndex(t, reflect.Value{})
			return true
		}
		return false
	}

	child := node.MapIndex(t)
	if !child.IsValid() {
		return false
	}
	child = reflect.ValueOf(child.Interface())
	if vl.deleteRecursive(child, tokens[1:]) {
		if child.Len() == 0 {
			node.SetMapIndex(t, reflect.Value{})
		}
		return true
	}

	return false
}
