package control

import (
	"github.com/siyul-park/uniflow/ext/pkg/language"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// AddToHook returns a function that adds hook to the provided hook.
func AddToHook() hook.Register {
	return hook.RegisterFunc(func(h *hook.Hook) error {
		h.AddLoadHook(symbol.LoadFunc(func(sb *symbol.Symbol) error {
			n := sb.Node
			if n, ok := n.(*BlockNode); ok {
				if err := n.Load(h); err != nil {
					return err
				}
			}
			return nil
		}))
		h.AddUnloadHook(symbol.UnloadFunc(func(sb *symbol.Symbol) error {
			n := sb.Node
			if n, ok := n.(*BlockNode); ok {
				if err := n.Unload(h); err != nil {
					return err
				}
			}
			return nil
		}))
		return nil
	})
}

// AddToScheme returns a function that adds node types and codecs to the provided spec.
func AddToScheme(module *language.Module, lang string) scheme.Register {
	return scheme.RegisterFunc(func(s *scheme.Scheme) error {
		expr, err := module.Load(lang)
		if err != nil {
			return err
		}

		s.AddKnownType(KindBlock, &BlockNodeSpec{})
		s.AddCodec(KindBlock, NewBlockNodeCodec(s))

		s.AddKnownType(KindCall, &CallNodeSpec{})
		s.AddCodec(KindCall, NewCallNodeCodec())

		s.AddKnownType(KindFork, &ForkNodeSpec{})
		s.AddCodec(KindFork, NewForkNodeCodec())

		s.AddKnownType(KindIf, &IfNodeSpec{})
		s.AddCodec(KindIf, NewIfNodeCodec(expr))

		s.AddKnownType(KindLoop, &LoopNodeSpec{})
		s.AddCodec(KindLoop, NewLoopNodeCodec())

		s.AddKnownType(KindMerge, &MergeNodeSpec{})
		s.AddCodec(KindMerge, NewMergeNodeCodec())

		s.AddKnownType(KindNOP, &NOPNodeSpec{})
		s.AddCodec(KindNOP, NewNOPNodeCodec())

		s.AddKnownType(KindReduce, &ReduceNodeSpec{})
		s.AddCodec(KindReduce, NewReduceNodeCodec(expr))

		s.AddKnownType(KindSession, &SessionNodeSpec{})
		s.AddCodec(KindSession, NewSessionNodeCodec())

		s.AddKnownType(KindSnippet, &SnippetNodeSpec{})
		s.AddCodec(KindSnippet, NewSnippetNodeCodec(module))

		s.AddKnownType(KindSplit, &SplitNodeSpec{})
		s.AddCodec(KindSplit, NewSplitNodeCodec())

		s.AddKnownType(KindSwitch, &SwitchNodeSpec{})
		s.AddCodec(KindSwitch, NewSwitchNodeCodec(expr))

		return nil
	})
}
