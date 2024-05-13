package network

import (
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// AddToHook returns a function that adds hook to the provided hook.
func AddToHook() func(*hook.Hook) error {
	return func(h *hook.Hook) error {
		h.AddLoadHook(symbol.LoadHookFunc(func(sym *symbol.Symbol) error {
			n := sym.Unwrap()
			if n, ok := n.(*HTTPServerNode); ok {
				return n.Listen()
			}
			return nil
		}))
		h.AddUnloadHook(symbol.UnloadHookFunc(func(sym *symbol.Symbol) error {
			n := sym.Unwrap()
			if n, ok := n.(*HTTPServerNode); ok {
				return n.Stop()
			}
			return nil
		}))
		return nil
	}
}

// AddToScheme returns a function that adds node types and codecs to the provided scheme.
func AddToScheme() func(*scheme.Scheme) error {
	return func(s *scheme.Scheme) error {
		s.AddKnownType(KindHTTPClient, &HTTPClientNodeSpec{})
		s.AddCodec(KindHTTPClient, NewHTTPClientNodeCodec())

		s.AddKnownType(KindHTTPServer, &HTTPServerNodeSpec{})
		s.AddCodec(KindHTTPServer, NewHTTPServerNodeCodec())

		s.AddKnownType(KindRoute, &RouteNodeSpec{})
		s.AddCodec(KindRoute, NewRouteNodeCodec())

		s.AddKnownType(KindWebSocketClient, &WebSocketClientNodeSpec{})
		s.AddCodec(KindWebSocketClient, NewWebSocketClientNodeCodec())

		s.AddKnownType(KindWebSocketUpgrade, &WebSocketUpgradeNodeSpec{})
		s.AddCodec(KindWebSocketUpgrade, NewWebSocketUpgradeNodeCodec())

		return nil
	}
}
