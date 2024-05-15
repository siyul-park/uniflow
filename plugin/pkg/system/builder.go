package system

import (
	"github.com/siyul-park/uniflow/pkg/event"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

type Config struct {
	Module *NativeModule
	Broker *event.Broker
}

// AddToHook returns a function that adds hook to the provided hook.
func AddToHook(config Config) func(*hook.Hook) error {
	broker := config.Broker

	return func(h *hook.Hook) error {
		if broker != nil {
			load := broker.Producer(TopicLoad)
			unload := broker.Producer(TopicUnload)

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
		}
		return nil
	}
}

// AddToScheme returns a function that adds node types and codecs to the provided scheme.
func AddToScheme(config Config) func(*scheme.Scheme) error {
	module := config.Module
	broker := config.Broker

	if module == nil {
		module = NewNativeModule()
	}

	return func(s *scheme.Scheme) error {
		s.AddKnownType(KindNative, &NativeNodeSpec{})
		s.AddCodec(KindNative, NewNativeNodeCodec(module))

		s.AddKnownType(KindTrigger, &TriggerNodeSpec{})
		s.AddCodec(KindTrigger, NewTriggerNodeCodec(broker))

		return nil
	}
}
