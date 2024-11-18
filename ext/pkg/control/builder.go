package control

import (
	"github.com/siyul-park/uniflow/ext/pkg/language"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
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

		definitions := []struct {
			kind  string
			codec scheme.Codec
			spec  spec.Spec
		}{
			{KindBlock, NewBlockNodeCodec(s), &BlockNodeSpec{}},
			{KindPipe, NewPipeNodeCodec(), &PipeNodeSpec{}},
			{KindFork, NewForkNodeCodec(), &ForkNodeSpec{}},
			{KindIf, NewIfNodeCodec(expr), &IfNodeSpec{}},
			{KindLoop, NewLoopNodeCodec(), &LoopNodeSpec{}},
			{KindMerge, NewMergeNodeCodec(), &MergeNodeSpec{}},
			{KindNOP, NewNOPNodeCodec(), &NOPNodeSpec{}},
			{KindReduce, NewReduceNodeCodec(expr), &ReduceNodeSpec{}},
			{KindRetry, NewRetryNodeCodec(), &RetryNodeSpec{}},
			{KindSession, NewSessionNodeCodec(), &SessionNodeSpec{}},
			{KindSnippet, NewSnippetNodeCodec(module), &SnippetNodeSpec{}},
			{KindSplit, NewSplitNodeCodec(), &SplitNodeSpec{}},
			{KindSwitch, NewSwitchNodeCodec(expr), &SwitchNodeSpec{}},
			{KindWait, NewWaitNodeCodec(), &WaitNodeSpec{}},
		}

		for _, def := range definitions {
			s.AddKnownType(def.kind, def.spec)
			s.AddCodec(def.kind, def.codec)
		}

		return nil
	})
}
