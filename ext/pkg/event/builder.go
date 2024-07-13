package event

import (
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// AddToHook returns a function that adds hook to the provided hook.
func AddToHook(broker *Broker) hook.Register {
	load := broker.Producer(TopicLoad)
	unload := broker.Producer(TopicUnload)

	return hook.RegisterFunc(func(h *hook.Hook) error {
		h.AddLoadHook(symbol.LoadFunc(func(sym *symbol.Symbol) error {
			n := sym.Node
			if n, ok := n.(*TriggerNode); ok {
				n.Listen()
			}
			return nil
		}))
		h.AddUnloadHook(symbol.UnloadFunc(func(sym *symbol.Symbol) error {
			n := sym.Node
			if n, ok := n.(*TriggerNode); ok {
				n.Shutdown()
			}
			return nil
		}))

		h.AddLoadHook(symbol.LoadFunc(func(sym *symbol.Symbol) error {
			e := New(sym.Spec)
			load.Produce(e)
			return nil
		}))
		h.AddUnloadHook(symbol.UnloadFunc(func(sym *symbol.Symbol) error {
			e := New(sym.Spec)
			unload.Produce(e)
			return nil
		}))
		return nil
	})
}

// AddToScheme returns a function that adds node types and codecs to the provided spec.
func AddToScheme(upsteam, downsteam *Broker) scheme.Register {
	return scheme.RegisterFunc(func(s *scheme.Scheme) error {
		s.AddKnownType(KindTrigger, &TriggerNodeSpec{})
		s.AddCodec(KindTrigger, NewTriggerNodeCodec(upsteam, downsteam))

		return nil
	})
}
