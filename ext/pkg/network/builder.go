package network

import (
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// AddToHook returns a function that adds hook to the provided hook.
func AddToHook() hook.Register {
	return hook.RegisterFunc(func(h *hook.Hook) error {
		h.AddLoadHook(symbol.LoadFunc(func(sb *symbol.Symbol) error {
			n := sb.Node
			if n, ok := n.(*HTTPListenNode); ok {
				return n.Listen()
			}
			return nil
		}))
		h.AddUnloadHook(symbol.UnloadFunc(func(sb *symbol.Symbol) error {
			n := sb.Node
			if n, ok := n.(*HTTPListenNode); ok {
				return n.Shutdown()
			}
			return nil
		}))
		return nil
	})
}

// AddToScheme returns a function that adds node types and codecs to the provided spec.
func AddToScheme() scheme.Register {
	return scheme.RegisterFunc(func(s *scheme.Scheme) error {
		s.AddKnownType(KindHTTP, &HTTPNodeSpec{})
		s.AddCodec(KindHTTP, NewHTTPNodeCodec())

		s.AddKnownType(KindListener, &ListenNodeSpec{})
		s.AddCodec(KindListener, NewListenNodeCodec())

		s.AddKnownType(KindProxy, &ProxyNodeSpec{})
		s.AddCodec(KindProxy, NewProxyNodeCodec())

		s.AddKnownType(KindRouter, &RouteNodeSpec{})
		s.AddCodec(KindRouter, NewRouteNodeCodec())

		s.AddKnownType(KindWebSocket, &WebSocketNodeSpec{})
		s.AddCodec(KindWebSocket, NewWebSocketNodeCodec())

		s.AddKnownType(KindGateway, &GatewayNodeSpec{})
		s.AddCodec(KindGateway, NewGatewayNodeCodec())

		return nil
	})
}
