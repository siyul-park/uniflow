package net

import (
	"github.com/siyul-park/uniflow/hook"
	"github.com/siyul-park/uniflow/scheme"
	"github.com/siyul-park/uniflow/symbol"
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
				return n.Shutdown()
			}
			return nil
		}))
		return nil
	}
}

// AddToScheme returns a function that adds node types and codecs to the provided spec.
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
