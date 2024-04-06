package control

import "github.com/siyul-park/uniflow/pkg/scheme"

// AddToScheme returns a function that adds node types and codecs to the provided scheme.
func AddToScheme() func(*scheme.Scheme) error {
	return func(s *scheme.Scheme) error {
		s.AddKnownType(KindCombine, &CombineNodeSpec{})
		s.AddCodec(KindCombine, NewCombineNodeCodec())

		s.AddKnownType(KindIterate, &IterateNodeSpec{})
		s.AddCodec(KindIterate, NewIterateNodeCodec())

		s.AddKnownType(KindJump, &JumpNodeSpec{})
		s.AddCodec(KindJump, NewJumpNodeCodec())

		s.AddKnownType(KindSnippet, &SnippetNodeSpec{})
		s.AddCodec(KindSnippet, NewSnippetNodeCodec())

		s.AddKnownType(KindSwitch, &SwitchNodeSpec{})
		s.AddCodec(KindSwitch, NewSwitchNodeCodec())

		return nil
	}
}
