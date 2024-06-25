package control

import (
	"github.com/siyul-park/uniflow/pkg/event"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/siyul-park/uniflow/plugin/pkg/language"
)

type Config struct {
	Broker     *event.Broker
	Module     *language.Module
	Expression string
}

// AddToHook returns a function that adds hook to the provided hook.
func AddToHook(config Config) func(*hook.Hook) error {
	broker := config.Broker

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
func AddToScheme(config Config) func(*scheme.Scheme) error {
	broker := config.Broker
	module := config.Module

	return func(s *scheme.Scheme) error {
		expr, err := module.Load(config.Expression)
		if err != nil {
			return err
		}

		s.AddKnownType(KindCall, &CallNodeSpec{})
		s.AddCodec(KindCall, NewCallNodeCodec())

		s.AddKnownType(KindIf, &IfNodeSpec{})
		s.AddCodec(KindIf, NewIfNodeCodec(expr))

		s.AddKnownType(KindLoop, &LoopNodeSpec{})
		s.AddCodec(KindLoop, NewLoopNodeCodec())

		s.AddKnownType(KindMerge, &MergeNodeSpec{})
		s.AddCodec(KindMerge, NewMergeNodeCodec())

		s.AddKnownType(KindNOP, &NOPNodeSpec{})
		s.AddCodec(KindNOP, NewNOPNodeCodec())

		s.AddKnownType(KindSnippet, &SnippetNodeSpec{})
		s.AddCodec(KindSnippet, NewSnippetNodeCodec(module))

		s.AddKnownType(KindSwitch, &SwitchNodeSpec{})
		s.AddCodec(KindSwitch, NewSwitchNodeCodec(expr))

		s.AddKnownType(KindTrigger, &TriggerNodeSpec{})
		s.AddCodec(KindTrigger, NewTriggerNodeCodec(broker))

		return nil
	}
}
