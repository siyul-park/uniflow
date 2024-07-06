package control

import (
	"github.com/siyul-park/uniflow/ext/pkg/language"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// AddToHook returns a function that adds hook to the provided hook.
func AddToHook() func(*hook.Hook) error {
	return func(h *hook.Hook) error {
		h.AddLoadHook(symbol.LoadFunc(func(sym *symbol.Symbol) error {
			n := sym.Unwrap()
			if n, ok := n.(*BlockNode); ok {
				nodes := n.Nodes()
				for i := len(nodes) - 1; i >= 0; i-- {
					n := nodes[i]
					sym, ok := n.(*symbol.Symbol)
					if !ok {
						sym = symbol.New(&spec.Meta{}, n)
					}
					if err := h.Load(sym); err != nil {
						return err
					}
				}
			}
			return nil
		}))
		h.AddUnloadHook(symbol.UnloadFunc(func(sym *symbol.Symbol) error {
			n := sym.Unwrap()
			if n, ok := n.(*BlockNode); ok {
				nodes := n.Nodes()
				for _, n := range nodes {
					sym, ok := n.(*symbol.Symbol)
					if !ok {
						sym = symbol.New(&spec.Meta{}, n)
					}
					if err := h.Unload(sym); err != nil {
						return err
					}
				}
			}
			return nil
		}))
		return nil
	}
}

// AddToScheme returns a function that adds node types and codecs to the provided spec.
func AddToScheme(module *language.Module, lang string) func(*scheme.Scheme) error {
	return func(s *scheme.Scheme) error {
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

		s.AddKnownType(KindSession, &SessionNodeSpec{})
		s.AddCodec(KindSession, NewSessionNodeCodec())

		s.AddKnownType(KindSnippet, &SnippetNodeSpec{})
		s.AddCodec(KindSnippet, NewSnippetNodeCodec(module))

		s.AddKnownType(KindSwitch, &SwitchNodeSpec{})
		s.AddCodec(KindSwitch, NewSwitchNodeCodec(expr))

		return nil
	}
}
