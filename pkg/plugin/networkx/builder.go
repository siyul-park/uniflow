package networkx

import (
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

func AddToHooks() func(*hook.Hook) error {
	return func(h *hook.Hook) error {
		h.AddLoadHook(symbol.LoadHookFunc(func(n node.Node) error {
			if n, ok := n.(*HTTPNode); ok {
				go func() { n.Start() }()
			}
			return nil
		}))
		return nil
	}
}

func AddToScheme() func(*scheme.Scheme) error {
	return func(s *scheme.Scheme) error {
		s.AddKnownType(KindHTTP, &HTTPSpec{})
		s.AddCodec(KindHTTP, scheme.CodecWithType[*HTTPSpec](func(spec *HTTPSpec) (node.Node, error) {
			return NewHTTPNode(HTTPNodeConfig{
				ID:      spec.ID,
				Address: spec.Address,
			}), nil
		}))

		s.AddKnownType(KindRouter, &RouterSpec{})
		s.AddCodec(KindRouter, scheme.CodecWithType[*RouterSpec](func(spec *RouterSpec) (node.Node, error) {
			n := NewRouterNode(RouterNodeConfig{
				ID: spec.ID,
			})
			for _, r := range spec.Routes {
				n.Add(r.Method, r.Path, r.Port)
			}
			return n, nil
		}))

		return nil
	}
}
