package sys

import "github.com/siyul-park/uniflow/scheme"

// AddToScheme returns a function that adds node types and codecs to the provided spec.
func AddToScheme(module *NativeModule) func(*scheme.Scheme) error {
	return func(s *scheme.Scheme) error {
		s.AddKnownType(KindNative, &NativeNodeSpec{})
		s.AddCodec(KindNative, NewNativeNodeCodec(module))

		return nil
	}
}
