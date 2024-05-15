package system

import (
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// AddToScheme returns a function that adds node types and codecs to the provided scheme.
func AddToScheme(table *NativeTable) func(*scheme.Scheme) error {
	if table == nil {
		table = NewNativeTable()
	}
	return func(s *scheme.Scheme) error {
		s.AddKnownType(KindNative, &NativeNodeSpec{})
		s.AddCodec(KindNative, NewNativeNodeCodec(table))

		return nil
	}
}
