package meta

import (
	"github.com/gofrs/uuid"
)

// Unstructured implements the Spec interface with a flexible key-value structure.
type Unstructured struct {
	// ID is the unique identifier of the node.
	ID uuid.UUID `json:"id,omitempty" yaml:"id,omitempty" validate:"required"`
	// Namespace groups nodes logically.
	Namespace string `json:"namespace" yaml:"namespace" validate:"required"`
	// Name is the human-readable name of the node.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Annotations hold additional metadata.
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	// Fields contain custom data in a flexible, key-value format.
	Fields map[string]any `json:",inline" yaml:",inline"`
}

// Key constants for commonly used fields in Unstructured.
const (
	KeyID          = "id"
	KeyNamespace   = "namespace"
	KeyName        = "name"
	KeyAnnotations = "annotations"
)

var _ Meta = (*Unstructured)(nil)

// GetID retrieves the ID of the node.
func (u *Unstructured) GetID() uuid.UUID {
	return u.ID
}

// SetID assigns the ID to the node.
func (u *Unstructured) SetID(val uuid.UUID) {
	u.ID = val
}

// GetNamespace retrieves the namespace of the node.
func (u *Unstructured) GetNamespace() string {
	return u.Namespace
}

// SetNamespace assigns the namespace to the node.
func (u *Unstructured) SetNamespace(val string) {
	u.Namespace = val
}

// GetName retrieves the name of the node.
func (u *Unstructured) GetName() string {
	return u.Name
}

// SetName assigns the name to the node.
func (u *Unstructured) SetName(val string) {
	u.Name = val
}

// GetAnnotations retrieves the annotations of the node.
func (u *Unstructured) GetAnnotations() map[string]string {
	return u.Annotations
}

// SetAnnotations assigns annotations to the node.
func (u *Unstructured) SetAnnotations(val map[string]string) {
	u.Annotations = val
}

// Get retrieves the value associated with the given key.
func (u *Unstructured) Get(key string) (any, bool) {
	switch key {
	case KeyID:
		return u.ID, true
	case KeyNamespace:
		return u.Namespace, true
	case KeyName:
		return u.Name, true
	case KeyAnnotations:
		return u.Annotations, true
	default:
		if u.Fields == nil {
			return nil, false
		}
		val, ok := u.Fields[key]
		return val, ok
	}
}

// Set assigns a value to the specified key.
func (u *Unstructured) Set(key string, val any) {
	switch key {
	case KeyID:
		if v, ok := val.(uuid.UUID); ok {
			u.ID = v
		}
	case KeyNamespace:
		if v, ok := val.(string); ok {
			u.Namespace = v
		}
	case KeyName:
		if v, ok := val.(string); ok {
			u.Name = v
		}
	case KeyAnnotations:
		if v, ok := val.(map[string]string); ok {
			u.Annotations = v
		}
	default:
		if u.Fields == nil {
			u.Fields = make(map[string]any)
		}
		u.Fields[key] = val
	}
}
