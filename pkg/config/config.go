package config

import (
	"reflect"
	"strings"
	"sync"
)

// Config represents the configuration of your application.
type Config struct {
	data map[string]any
	mu   sync.RWMutex
}

var typeAny = reflect.TypeOf((*any)(nil)).Elem()

// New creates a new configuration instance.
func New() *Config {
	return &Config{
		data: make(map[string]any),
	}
}

// Set sets the value for the specified key in the configuration.
func (c *Config) Set(key string, val any) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	tokens := strings.Split(key, ".")
	v := reflect.ValueOf(val)

	node := reflect.ValueOf(c.data)
	for i, token := range tokens {
		t := reflect.ValueOf(token)

		if node.Kind() != reflect.Map || node.Type().Key().Kind() != reflect.String {
			return false
		}

		if i == len(tokens)-1 {
			node.SetMapIndex(t, v)
			return true
		}

		child := node.MapIndex(t)
		if !child.IsValid() {
			child = reflect.MakeMap(reflect.MapOf(reflect.TypeOf(""), typeAny))
			node.SetMapIndex(t, child)
		}
		node = reflect.ValueOf(child.Interface())
	}
	return false
}

// Get retrieves the value for the specified key from the configuration.
// If the key is not present, it returns a default value and a boolean indicating the key's existence.
func (c *Config) Get(key string) (any, bool) {
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

// Delete removes the specified key from the configuration.
func (c *Config) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	tokens := strings.Split(key, ".")
	return c.deleteRecursive(reflect.ValueOf(c.data), tokens)
}

// deleteRecursive is a recursive helper function for Delete.
func (c *Config) deleteRecursive(node reflect.Value, tokens []string) bool {
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
