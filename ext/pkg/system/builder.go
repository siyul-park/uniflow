package system

import (
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// AddToScheme returns a function that adds node types and codecs to the provided spec.
func AddToScheme(table *NativeTable) scheme.Register {
	return scheme.RegisterFunc(func(s *scheme.Scheme) error {
		definitions := []struct {
			kind  string
			codec scheme.Codec
			spec  spec.Spec
		}{
			{KindNative, NewNativeNodeCodec(table), &NativeNodeSpec{}},
		}

		for _, def := range definitions {
			s.AddKnownType(def.kind, def.spec)
			s.AddCodec(def.kind, def.codec)
		}

		return nil
	})
}
