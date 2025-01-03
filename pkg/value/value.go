package value

import (
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/resource"
)

// Value defines the interface for a value with various attributes.
type Value struct {
	// ID is the unique identifier of the value.
	ID uuid.UUID `json:"id" bson:"_id" yaml:"id" map:"id" validate:"required"`
	// Namespace groups values logically.
	Namespace string `json:"namespace" bson:"namespace" yaml:"namespace" map:"namespace" validate:"required"`
	// Name is the human-readable name of the value.
	Name string `json:"name,omitempty" bson:"name,omitempty" yaml:"name,omitempty" map:"name,omitempty"`
	// Annotations hold additional metadata.
	Annotations map[string]string `json:"annotations,omitempty" bson:"annotations,omitempty" yaml:"annotations,omitempty" map:"annotations,omitempty"`
	// Data holds the value's actual data.
	Data any `json:"data" bson:"data" yaml:"data" map:"data" validate:"required"`
}

// Key constants for commonly used fields.
const (
	KeyID          = "id"
	KeyNamespace   = "namespace"
	KeyName        = "name"
	KeyAnnotations = "annotations"
	KeyData        = "data"
)

var _ resource.Resource = (*Value)(nil)

// New creates and returns a new instance of Value.
func New() *Value {
	return &Value{}
}

// GetID returns the value's unique identifier.
func (s *Value) GetID() uuid.UUID {
	return s.ID
}

// SetID assigns a unique identifier to the value.
func (s *Value) SetID(val uuid.UUID) {
	s.ID = val
}

// GetNamespace returns the value's namespace.
func (s *Value) GetNamespace() string {
	return s.Namespace
}

// SetNamespace sets the value's namespace.
func (s *Value) SetNamespace(val string) {
	s.Namespace = val
}

// GetName returns the value's name.
func (s *Value) GetName() string {
	return s.Name
}

// SetName sets the value's name.
func (s *Value) SetName(val string) {
	s.Name = val
}

// GetAnnotations returns the value's annotations.
func (s *Value) GetAnnotations() map[string]string {
	return s.Annotations
}

// SetAnnotations sets the value's annotations.
func (s *Value) SetAnnotations(val map[string]string) {
	s.Annotations = val
}

// GetData returns the value's data.
func (s *Value) GetData() any {
	return s.Data
}

// SetData sets the value's data.
func (s *Value) SetData(val any) {
	s.Data = val
}

// IsIdentified checks whether the Value instance has a unique identifier or name.
func (s *Value) IsIdentified() bool {
	return s.ID != uuid.Nil || s.Name != ""
}
