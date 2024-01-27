package network

import (
	"context"
	"time"

	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// AddToHooks returns a function that adds hook to the provided hook.
func AddToHooks(ctx context.Context, timeout time.Duration) func(*hook.Hook) error {
	return func(h *hook.Hook) error {
		h.AddLoadHook(symbol.LoadHookFunc(func(n node.Node) error {
			if n, ok := n.(*HTTPNode); ok {
				ctx, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()

				return n.Listen(ctx)
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

		return nil
	}
}
