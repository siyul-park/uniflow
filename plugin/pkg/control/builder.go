package control

import "github.com/siyul-park/uniflow/pkg/scheme"

// AddToScheme returns a function that adds node types and codecs to the provided scheme.
func AddToScheme() func(*scheme.Scheme) error {
	return func(s *scheme.Scheme) error {
		s.AddKnownType(KindCall, &CallNodeSpec{})
		s.AddCodec(KindCall, NewCallNodeCodec())

		s.AddKnownType(KindForEach, &ForEachNodeSpec{})
		s.AddCodec(KindForEach, NewForEachNodeCodec())

		s.AddKnownType(KindMerge, &MergeNodeSpec{})
		s.AddCodec(KindMerge, NewMergeNodeCodec())

		s.AddKnownType(KindNoOp, &NoOpNodeSpec{})
		s.AddCodec(KindNoOp, NewNoOpNodeCodec())

		s.AddKnownType(KindSnippet, &SnippetNodeSpec{})
		s.AddCodec(KindSnippet, NewSnippetNodeCodec())

		s.AddKnownType(KindSwitch, &SwitchNodeSpec{})
		s.AddCodec(KindSwitch, NewSwitchNodeCodec())

		return nil
	}
}
