package event

import (
	"context"

	"github.com/siyul-park/uniflow/pkg/boot"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// AddToHook returns a function that adds hook to the provided hook.
func AddToHook(upsteam, downsteam *Broker) func(*hook.Hook) error {
	boots := upsteam.Producer(TopicBoot)
	loads := upsteam.Producer(TopicLoad)
	unloads := upsteam.Producer(TopicUnload)

	return func(h *hook.Hook) error {
		h.AddLoadHook(symbol.LoadHookFunc(func(sym *symbol.Symbol) error {
			n := sym.Unwrap()
			if n, ok := n.(*TriggerNode); ok {
				n.Listen()
			}
			return nil
		}))
		h.AddUnloadHook(symbol.UnloadHookFunc(func(sym *symbol.Symbol) error {
			n := sym.Unwrap()
			if n, ok := n.(*TriggerNode); ok {
				n.Shutdown()
			}
			return nil
		}))

		h.AddBootHook(boot.BootHookFunc(func(ctx context.Context) error {
			e := New(nil)
			boots.Produce(e)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-e.Done():
				return nil
			}
		}))
		h.AddLoadHook(symbol.LoadHookFunc(func(sym *symbol.Symbol) error {
			e := New(sym.Spec())
			loads.Produce(e)
			return nil
		}))
		h.AddUnloadHook(symbol.UnloadHookFunc(func(sym *symbol.Symbol) error {
			e := New(sym.Spec())
			unloads.Produce(e)
			return nil
		}))
		return nil
	}
}

// AddToScheme returns a function that adds node types and codecs to the provided spec.
func AddToScheme(upsteam, downsteam *Broker) func(*scheme.Scheme) error {
	return func(s *scheme.Scheme) error {
		s.AddKnownType(KindTrigger, &TriggerNodeSpec{})
		s.AddCodec(KindTrigger, NewTriggerNodeCodec(upsteam, downsteam))

		return nil
	}
}
