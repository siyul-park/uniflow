package scheme

import (
	"github.com/gofrs/uuid"
	"reflect"
	"slices"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// Scheme manages type information and decodes spec implementations into node objects within a workflow environment.
type Scheme struct {
	types  map[string]reflect.Type
	codecs map[string]Codec
	mu     sync.RWMutex
}

var _ Codec = (*Scheme)(nil)

// New creates a new Scheme instance with initialized type and codec maps.
func New() *Scheme {
	return &Scheme{
		types:  make(map[string]reflect.Type),
		codecs: make(map[string]Codec),
	}
}

// Kinds returns all unique kinds from types and codecs.
func (s *Scheme) Kinds() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	kinds := make([]string, 0, len(s.types))
	for kind := range s.types {
		kinds = append(kinds, kind)
	}
	for kind := range s.codecs {
		if !slices.Contains(kinds, kind) {
			kinds = append(kinds, kind)
		}
	}
	return kinds
}

// AddKnownType associates a spec type with a kind and returns true if successful.
func (s *Scheme) AddKnownType(kind string, sp spec.Spec) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sp == nil {
		return false
	}
	if _, ok := s.types[kind]; ok {
		return false
	}
	s.types[kind] = reflect.TypeOf(sp)
	return true
}

// RemoveKnownType removes the spec type associated with a kind.
func (s *Scheme) RemoveKnownType(kind string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.types[kind]; !ok {
		return false
	}
	delete(s.types, kind)
	return true
}

// KnownType retrieves the type of the spec associated with the given kind.
func (s *Scheme) KnownType(kind string) reflect.Type {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.types[kind]
}

// AddCodec associates a codec with a specific kind and returns true if successful.
func (s *Scheme) AddCodec(kind string, codec Codec) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.codecs[kind]; ok {
		return false
	}
	s.codecs[kind] = codec
	return true
}

// RemoveCodec removes the codec associated with a kind.
func (s *Scheme) RemoveCodec(kind string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.codecs[kind]; !ok {
		return false
	}
	delete(s.codecs, kind)
	return true
}

// Codec retrieves the codec associated with the given kind.
func (s *Scheme) Codec(kind string) Codec {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.codecs[kind]
}

// Decode converts the input spec.Spec into a registered structured type if one exists.
func (s *Scheme) Decode(sp spec.Spec) (spec.Spec, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	typ, ok := s.types[sp.GetKind()]
	if !ok {
		return sp, nil
	}

	value := reflect.New(typ).Elem()
	if value.Kind() == reflect.Pointer {
		value.Set(reflect.New(typ.Elem()))
	}

	structured, ok := value.Interface().(spec.Spec)
	if !ok {
		return sp, nil
	}

	if err := spec.Convert(sp, structured); err != nil {
		return nil, err
	}

	if structured.GetID() == uuid.Nil {
		structured.SetID(uuid.Must(uuid.NewV7()))
	}
	return structured, nil
}

// Compile decodes the given spec into a node using the associated codec.
func (s *Scheme) Compile(sp spec.Spec) (node.Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cdc := s.Codec(sp.GetKind())
	if cdc == nil {
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	}
	return cdc.Compile(sp)
}
