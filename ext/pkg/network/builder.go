package network

import (
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// AddToHook returns a function that adds hook to the provided hook.
func AddToHook() hook.Register {
	return hook.RegisterFunc(func(h *hook.Hook) error {
		h.AddLoadHook(symbol.LoadFunc(func(sb *symbol.Symbol) error {
			var n *HTTPListenNode
			if node.As(sb, &n) {
				return n.Listen()
			}
			return nil
		}))
		h.AddUnloadHook(symbol.UnloadFunc(func(sb *symbol.Symbol) error {
			var n *HTTPListenNode
			if node.As(sb, &n) {
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
		definitions := []struct {
			kind  string
			codec scheme.Codec
			spec  spec.Spec
		}{
			{KindHTTP, NewHTTPNodeCodec(), &HTTPNodeSpec{}},
			{KindListener, NewListenNodeCodec(), &ListenNodeSpec{}},
			{KindRouter, NewRouteNodeCodec(), &RouteNodeSpec{}},
			{KindWebSocket, NewWebSocketNodeCodec(), &WebSocketNodeSpec{}},
			{KindGateway, NewGatewayNodeCodec(), &GatewayNodeSpec{}},
		}

		for _, def := range definitions {
			s.AddKnownType(def.kind, def.spec)
			s.AddCodec(def.kind, def.codec)
		}

		return nil
	})
}
