package scheme

import (
	"reflect"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/template"
	"github.com/siyul-park/uniflow/pkg/types"
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

// AddKnownType associates a spec type with a kind and returns true if successful.
func (s *Scheme) AddKnownType(kind string, sp spec.Spec) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

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

	if _, ok := s.types[kind]; ok {
		delete(s.types, kind)
		return true
	}
	return false
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

	if _, ok := s.codecs[kind]; ok {
		delete(s.codecs, kind)
		return true
	}
	return false
}

// Codec retrieves the codec associated with the given kind.
func (s *Scheme) Codec(kind string) Codec {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.codecs[kind]
}

// Compile decodes the given spec into node using the associated codec.
func (s *Scheme) Compile(sp spec.Spec) (node.Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	codec := s.Codec(sp.GetKind())
	if codec == nil {
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	}
	return codec.Compile(sp)
}

// IsBound checks if the spec is bound to any of the provided secrets.
func (s *Scheme) IsBound(sp spec.Spec, secrets ...*secret.Secret) bool {
	for _, vals := range sp.GetEnv() {
		for _, val := range vals {
			examples := make([]*secret.Secret, 0, 2)
			if val.ID != uuid.Nil {
				examples = append(examples, &secret.Secret{ID: val.ID})
			}
			if val.Name != "" {
				examples = append(examples, &secret.Secret{Namespace: sp.GetNamespace(), Name: val.Name})
			}

			for _, sec := range secrets {
				if len(resource.Match(sec, examples...)) > 0 {
					return true
				}
			}
		}
	}
	return false
}

// Bind processes the environment variables in the spec using the provided secrets.
func (s *Scheme) Bind(sp spec.Spec, secrets ...*secret.Secret) (spec.Spec, error) {
	doc, err := types.Marshal(sp)
	if err != nil {
		return nil, err
	}

	unstructured := &spec.Unstructured{}
	if err := types.Unmarshal(doc, unstructured); err != nil {
		return nil, err
	}

	env := map[string]any{}
	for key, vals := range unstructured.GetEnv() {
		for i, val := range vals {
			if val.ID != uuid.Nil || val.Name != "" {
				example := &secret.Secret{
					ID:        val.ID,
					Namespace: unstructured.GetNamespace(),
					Name:      val.Name,
				}

				var sec *secret.Secret
				for _, s := range secrets {
					if len(resource.Match(s, example)) > 0 {
						sec = s
						break
					}
				}
				if sec == nil {
					continue
				}

				v, err := template.Execute(val.Value, sec.Data)
				if err != nil {
					return nil, err
				}

				val.ID = sec.GetID()
				val.Name = sec.GetName()
				val.Value = v

				vals[i] = val
			}

			env[key] = val.Value
		}

		if _, ok := env[key]; !ok {
			return nil, errors.WithStack(encoding.ErrUnsupportedValue)
		}
	}

	if len(env) > 0 {
		fields, err := template.Execute(unstructured.Fields, env)
		if err != nil {
			return nil, err
		}
		unstructured.Fields = fields.(map[string]any)
	}

	return unstructured, nil
}

// Decode converts the provided spec.Spec into a structured representation using reflection and encoding utilities.
func (s *Scheme) Decode(sp spec.Spec) (spec.Spec, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, err := types.Marshal(sp)
	if err != nil {
		return nil, err
	}

	typ, ok := s.types[sp.GetKind()]
	if !ok {
		return sp, nil
	}

	val := reflect.New(typ).Elem()
	if val.Kind() == reflect.Pointer {
		val.Set(reflect.New(typ.Elem()))
	}

	structured, ok := val.Interface().(spec.Spec)
	if !ok {
		return sp, nil
	}

	if err := types.Unmarshal(doc, structured); err != nil {
		return nil, err
	}

	return structured, nil
}
