package systemx

import (
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
)

// AddToScheme returns a function that adds types and codecs to the provided scheme.
func AddToScheme(storage *storage.Storage) func(*scheme.Scheme) error {
	return func(s *scheme.Scheme) error {
		s.AddKnownType(KindReflect, &ReflectSpec{})
		s.AddCodec(KindReflect, scheme.CodecWithType[*ReflectSpec](func(spec *ReflectSpec) (node.Node, error) {
			return NewReflectNode(ReflectNodeConfig{
				OP:      spec.OP,
				Storage: storage,
			}), nil
		}))

		return nil
	}
}
