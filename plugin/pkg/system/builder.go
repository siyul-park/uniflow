package system

import (
	"github.com/siyul-park/uniflow/pkg/spec"
)

// AddToScheme returns a function that adds node types and codecs to the provided spec.
func AddToScheme(module *NativeModule) func(*spec.Scheme) error {
	return func(s *spec.Scheme) error {
		s.AddKnownType(KindNative, &NativeNodeSpec{})
		s.AddCodec(KindNative, NewNativeNodeCodec(module))

		return nil
	}
}
