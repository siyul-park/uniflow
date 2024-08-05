package system

import "github.com/siyul-park/uniflow/pkg/scheme"

// AddToScheme returns a function that adds node types and codecs to the provided spec.
func AddToScheme(table *NativeTable) scheme.Register {
	return scheme.RegisterFunc(func(s *scheme.Scheme) error {
		s.AddKnownType(KindNative, &NativeNodeSpec{})
		s.AddCodec(KindNative, NewNativeNodeCodec(table))

		return nil
	})
}
