package secret

import (
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/resource"
)

// Secret defines the interface for a secret with various attributes.
type Secret struct {
	// ID is the unique identifier of the secret.
	ID uuid.UUID `json:"id,omitempty" bson:"_id,omitempty" yaml:"id,omitempty" map:"id,omitempty"`
	// Namespace groups secrets logically.
	Namespace string `json:"namespace,omitempty" bson:"namespace,omitempty" yaml:"namespace,omitempty" map:"namespace,omitempty"`
	// Name is the human-readable name of the secret.
	Name string `json:"name,omitempty" bson:"name,omitempty" yaml:"name,omitempty" map:"name,omitempty"`
	// Annotations hold additional metadata.
	Annotations map[string]string `json:"annotations,omitempty" bson:"annotations,omitempty" yaml:"annotations,omitempty" map:"annotations,omitempty"`
	// Data holds the secret's actual data.
	Data any `json:"data" bson:"data" yaml:"data" map:"data"`
}

// Key constants for commonly used fields.
const (
	KeyID          = "id"
	KeyNamespace   = "namespace"
	KeyName        = "name"
	KeyAnnotations = "annotations"
	KeyData        = "data"
)

var _ resource.Resource = (*Secret)(nil)

// New creates and returns a new instance of Secret.
func New() *Secret {
	return &Secret{}
}

// GetID returns the secret's unique identifier.
func (s *Secret) GetID() uuid.UUID {
	return s.ID
}

// SetID assigns a unique identifier to the secret.
func (s *Secret) SetID(val uuid.UUID) {
	s.ID = val
}

// GetNamespace returns the secret's namespace.
func (s *Secret) GetNamespace() string {
	return s.Namespace
}

// SetNamespace sets the secret's namespace.
func (s *Secret) SetNamespace(val string) {
	s.Namespace = val
}

// GetName returns the secret's name.
func (s *Secret) GetName() string {
	return s.Name
}

// SetName sets the secret's name.
func (s *Secret) SetName(val string) {
	s.Name = val
}

// GetAnnotations returns the secret's annotations.
func (s *Secret) GetAnnotations() map[string]string {
	return s.Annotations
}

// SetAnnotations sets the secret's annotations.
func (s *Secret) SetAnnotations(val map[string]string) {
	s.Annotations = val
}

// GetData returns the secret's data.
func (s *Secret) GetData() any {
	return s.Data
}

// SetData sets the secret's data.
func (s *Secret) SetData(val any) {
	s.Data = val
}
