package network

import (
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// AddToHook returns a function that adds hook to the provided hook.
func AddToHook() func(*hook.Hook) error {
	return func(h *hook.Hook) error {
		h.AddLoadHook(symbol.LoadHookFunc(func(n node.Node) error {
			if n, ok := n.(*HTTPNode); ok {
				return n.Listen()
			}
			return nil
		}))
		return nil
	}
}

// AddToScheme returns a function that adds node types and codecs to the provided scheme.
func AddToScheme() func(*scheme.Scheme) error {
	return func(s *scheme.Scheme) error {
		s.AddKnownType(KindHTTP, &HTTPNodeSpec{})
		s.AddCodec(KindHTTP, NewHTTPNodeCodec())

		s.AddKnownType(KindRoute, &RouteNodeSpec{})
		s.AddCodec(KindRoute, NewRouteNodeCodec())

		return nil
	}
}
