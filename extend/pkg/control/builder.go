package control

import (
	"github.com/siyul-park/uniflow/extend/pkg/language"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// AddToScheme returns a function that adds node types and codecs to the provided spec.
func AddToScheme(module *language.Module, lang string) func(*scheme.Scheme) error {
	return func(s *scheme.Scheme) error {
		expr, err := module.Load(lang)
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

		return nil
	}
}
