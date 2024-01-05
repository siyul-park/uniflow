package control

import "github.com/siyul-park/uniflow/pkg/scheme"

func Schemes() func(*scheme.Scheme) error {
	return func(s *scheme.Scheme) error {
		s.AddKnownType(KindSnippet, &SnippetNodeSpec{})
		s.AddCodec(KindSnippet, NewSnippetNodeCodec())

		return nil
	}
}
