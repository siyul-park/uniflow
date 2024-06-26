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
			if n, ok := n.(*HTTPListenNode); ok {
				return n.Listen()
			}
			return nil
		}))
		h.AddUnloadHook(symbol.UnloadHookFunc(func(sym *symbol.Symbol) error {
			n := sym.Unwrap()
			if n, ok := n.(*HTTPListenNode); ok {
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
		s.AddKnownType(KindHTTP, &HTTPNodeSpec{})
		s.AddCodec(KindHTTP, NewHTTPNodeCodec())

		s.AddKnownType(KindListen, &ListenNodeSpec{})
		s.AddCodec(KindListen, NewListenNodeCodec())

		s.AddKnownType(KindRoute, &RouteNodeSpec{})
		s.AddCodec(KindRoute, NewRouteNodeCodec())

		s.AddKnownType(KindWebSocket, &WebSocketNodeSpec{})
		s.AddCodec(KindWebSocket, NewWebSocketNodeCodec())

		s.AddKnownType(KindUpgrade, &UpgradeNodeSpec{})
		s.AddCodec(KindUpgrade, NewUpgradeNodeCodec())

		return nil
	}
}
