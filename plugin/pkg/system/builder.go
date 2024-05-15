package system

import (
	"github.com/siyul-park/uniflow/pkg/scheme"
)

type Config struct {
	Module *NativeModule
}

// AddToScheme returns a function that adds node types and codecs to the provided scheme.
func AddToScheme(config Config) func(*scheme.Scheme) error {
	module := config.Module
	if module == nil {
		module = NewNativeModule()
	}
	return func(s *scheme.Scheme) error {
		s.AddKnownType(KindNative, &NativeNodeSpec{})
		s.AddCodec(KindNative, NewNativeNodeCodec(module))

		return nil
	}
}
