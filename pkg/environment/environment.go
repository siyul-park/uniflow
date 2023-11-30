package config

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// Environment represents the environment of your application.
type Environment struct {
	data map[string]any
	mu   sync.RWMutex
}

var typeAny = reflect.TypeOf((*any)(nil)).Elem()

// New creates a new environment instance.
func New() *Environment {
	return &Environment{
		data: make(map[string]any),
	}
}

// Set sets the value for the specified key in the environment.
func (c *Environment) Set(key string, val any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	tokens := strings.Split(key, ".")
	v := reflect.ValueOf(val)

	node := reflect.ValueOf(c.data)
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
func (c *Environment) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tokens := strings.Split(key, ".")

	node := reflect.ValueOf(c.data)
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
func (c *Environment) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tokens := strings.Split(key, ".")
	c.deleteRecursive(reflect.ValueOf(c.data), tokens)
}

// GetData returns a read-only copy of the environment data.
func (c *Environment) GetData() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]any, len(c.data))
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

func (c *Environment) deleteRecursive(node reflect.Value, tokens []string) bool {
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
	if c.deleteRecursive(child, tokens[1:]) {
		if child.Len() == 0 {
			node.SetMapIndex(t, reflect.Value{})
		}
		return true
	}

	return false
}
