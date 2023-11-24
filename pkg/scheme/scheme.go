package scheme

import (
	"reflect"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/node"
)

type (
	// Scheme defines methods for decode Spec.
	Scheme struct {
		types  map[string]reflect.Type
		codecs map[string]Codec
		mu     sync.RWMutex
	}
)

var _ Codec = &Scheme{}

// New returns a new Scheme.
func New() *Scheme {
	return &Scheme{
		types:  make(map[string]reflect.Type),
		codecs: make(map[string]Codec),
	}
}

// AddKnownType adds a new Type and Spec to the Scheme.
func (s *Scheme) AddKnownType(kind string, spec Spec) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.types[kind] = reflect.TypeOf(spec)
}

// KnownType returns the reflect.Type of the Spec with the given kind.
func (s *Scheme) KnownType(kind string) (reflect.Type, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, ok := s.types[kind]
	return t, ok
}

// AddCodec adds a new Codec to the Scheme.
func (s *Scheme) AddCodec(kind string, codec Codec) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.codecs[kind] = codec
}

// Codec returns Codec with the given kind.
func (s *Scheme) Codec(kind string) (Codec, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, ok := s.codecs[kind]
	return c, ok
}

// New returns a new Spec with the given kind.
func (s *Scheme) New(kind string) (Spec, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if t, ok := s.types[kind]; !ok {
		return nil, false
	} else {
		zero := reflect.New(t)
		if zero.Elem().Kind() == reflect.Pointer {
			zero.Elem().Set(reflect.New(t.Elem()))
		}
		v, ok := zero.Elem().Interface().(Spec)
		return v, ok
	}
}

// Decode decodes the given Spec into a node.Node.
func (s *Scheme) Decode(spec Spec) (node.Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	kind := spec.GetKind()
	if kind == "" {
		if kinds := s.Kinds(spec); len(kinds) > 0 {
			kind = kinds[0]
		}
	}

	if unstructured, ok := spec.(*Unstructured); ok {
		if structured, ok := s.New(kind); ok {
			if err := unstructured.Unmarshal(structured); err != nil {
				return nil, err
			} else {
				spec = structured
			}
		}
	}

	if codec, ok := s.codecs[kind]; ok {
		return codec.Decode(spec)
	}
	return nil, errors.WithStack(encoding.ErrUnsupportedValue)
}

// Kinds returns the kinds of the given Spec.
func (s *Scheme) Kinds(spec Spec) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	typ := reflect.TypeOf(spec)

	var kinds []string
	for kind, t := range s.types {
		if t == typ {
			kinds = append(kinds, kind)
		}
	}

	return kinds
}
