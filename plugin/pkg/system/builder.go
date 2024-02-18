package system

import (
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// AddToScheme returns a function that adds node types and codecs to the provided scheme.
func AddToScheme(table *BridgeTable) func(*scheme.Scheme) error {
	if table == nil {
		table = NewBridgeTable()
	}

	return func(s *scheme.Scheme) error {
		s.AddKnownType(KindBridge, &BridgeNodeSpec{})
		s.AddCodec(KindBridge, NewBridgeNodeCodec(table))

		return nil
	}
}
