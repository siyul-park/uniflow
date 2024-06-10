package control

import (
	"github.com/siyul-park/uniflow/pkg/event"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// AddToHook returns a function that adds hook to the provided hook.
func AddToHook(broker *event.Broker) func(*hook.Hook) error {
	load := broker.Producer(TopicLoad)
	unload := broker.Producer(TopicUnload)

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

		h.AddLoadHook(symbol.LoadHookFunc(func(sym *symbol.Symbol) error {
			e := event.New(sym.Spec())
			load.Produce(e)
			return nil
		}))
		h.AddUnloadHook(symbol.UnloadHookFunc(func(sym *symbol.Symbol) error {
			e := event.New(sym.Spec())
			unload.Produce(e)
			return nil
		}))
		return nil
	}
}

// AddToScheme returns a function that adds node types and codecs to the provided spec.
func AddToScheme(broker *event.Broker) func(*scheme.Scheme) error {
	return func(s *scheme.Scheme) error {
		s.AddKnownType(KindCall, &CallNodeSpec{})
		s.AddCodec(KindCall, NewCallNodeCodec())

		s.AddKnownType(KindIf, &IfNodeSpec{})
		s.AddCodec(KindIf, NewIfNodeCodec())

		s.AddKnownType(KindLoop, &LoopNodeSpec{})
		s.AddCodec(KindLoop, NewLoopNodeCodec())

		s.AddKnownType(KindMerge, &MergeNodeSpec{})
		s.AddCodec(KindMerge, NewMergeNodeCodec())

		s.AddKnownType(KindNOP, &NOPNodeSpec{})
		s.AddCodec(KindNOP, NewNOPNodeCodec())

		s.AddKnownType(KindSnippet, &SnippetNodeSpec{})
		s.AddCodec(KindSnippet, NewSnippetNodeCodec())

		s.AddKnownType(KindSwitch, &SwitchNodeSpec{})
		s.AddCodec(KindSwitch, NewSwitchNodeCodec())

		s.AddKnownType(KindTrigger, &TriggerNodeSpec{})
		s.AddCodec(KindTrigger, NewTriggerNodeCodec(broker))

		return nil
	}
}
