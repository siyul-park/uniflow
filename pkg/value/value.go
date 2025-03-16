package value

import (
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/meta"
)

// Value defines the interface for a value with various attributes.
type Value struct {
	// ID is the unique identifier of the value.
	ID uuid.UUID `json:"id,omitempty" yaml:"id,omitempty" validate:"required"`
	// Namespace groups values logically.
	Namespace string `json:"namespace" yaml:"namespace" validate:"required"`
	// Name is the human-readable name of the value.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Annotations hold additional metadata.
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	// Data holds the value's actual data.
	Data any `json:"data" yaml:"data" validate:"required"`
}

// Key constants for commonly used fields.
const (
	KeyID          = "id"
	KeyNamespace   = "namespace"
	KeyName        = "name"
	KeyAnnotations = "annotations"
	KeyData        = "data"
)

var _ meta.Meta = (*Value)(nil)

// New creates and returns a new instance of Value.
func New() *Value {
	return &Value{}
}

// GetID returns the value's unique identifier.
func (v *Value) GetID() uuid.UUID {
	return v.ID
}

// SetID assigns a unique identifier to the value.
func (v *Value) SetID(val uuid.UUID) {
	v.ID = val
}

// GetNamespace returns the value's namespace.
func (v *Value) GetNamespace() string {
	return v.Namespace
}

// SetNamespace sets the value's namespace.
func (v *Value) SetNamespace(val string) {
	v.Namespace = val
}

// GetName returns the value's name.
func (v *Value) GetName() string {
	return v.Name
}

// SetName sets the value's name.
func (v *Value) SetName(val string) {
	v.Name = val
}

// GetAnnotations returns the value's annotations.
func (v *Value) GetAnnotations() map[string]string {
	return v.Annotations
}

// SetAnnotations sets the value's annotations.
func (v *Value) SetAnnotations(val map[string]string) {
	v.Annotations = val
}

// GetData returns the value's data.
func (v *Value) GetData() any {
	return v.Data
}

// SetData sets the value's data.
func (v *Value) SetData(val any) {
	v.Data = val
}

// IsIdentified checks whether the Value instance has a unique identifier or name.
func (v *Value) IsIdentified() bool {
	return v.ID != uuid.Nil || v.Name != ""
}

// Is checks whether all non-zero fields in the target exist in the source with matching values.
func (v *Value) Is(val *Value) bool {
	if val.GetID() != uuid.Nil && val.GetID() != v.GetID() {
		return false
	}
	if val.GetNamespace() != "" && val.GetNamespace() != v.GetNamespace() {
		return false
	}
	if val.GetName() != "" && val.GetName() != v.GetName() {
		return false
	}
	return true
}
