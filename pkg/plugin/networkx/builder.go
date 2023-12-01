package networkx

import (
	"context"
	"time"

	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// AddToHooks returns a function that adds hooks for HTTPNode to the given hook.Hook.
func AddToHooks() func(*hook.Hook) error {
	return func(h *hook.Hook) error {
		h.AddLoadHook(symbol.LoadHookFunc(func(n node.Node) error {
			if n, ok := n.(*HTTPNode); ok {
				errChan := make(chan error)

				go func() {
					if err := n.Serve(); err != nil {
						errChan <- err
					}
				}()

				return n.WaitForListen(errChan)
			}
			return nil
		}))

		h.AddUnloadHook(symbol.UnloadHookFunc(func(n node.Node) error {
			if n, ok := n.(*HTTPNode); ok {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				return n.Shutdown(ctx)
			}
			return nil
		}))

		return nil
	}
}

// AddToScheme returns a function that adds schemes for HTTPNode and RouterNode to the given scheme.Scheme.
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
