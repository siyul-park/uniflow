package controllx

import (
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// AddToScheme returns a function that adds types and codecs to the provided scheme.
func AddToScheme() func(*scheme.Scheme) error {
	return func(s *scheme.Scheme) error {
		s.AddKnownType(KindSnippet, &SnippetSpec{})
		s.AddCodec(KindSnippet, scheme.CodecWithType[*SnippetSpec](func(spec *SnippetSpec) (node.Node, error) {
			return NewSnippetNode(SnippetNodeConfig{
				ID:   spec.ID,
				Lang: spec.Lang,
				Code: spec.Code,
			})
		}))

		s.AddKnownType(KindSwitch, &SwitchSpec{})
		s.AddCodec(KindSwitch, scheme.CodecWithType[*SwitchSpec](func(spec *SwitchSpec) (node.Node, error) {
			n := NewSwitchNode(SwitchNodeConfig{
				ID: spec.ID,
			})
			for _, v := range spec.Match {
				if err := n.Add(v.When, v.Port); err != nil {
					_ = n.Close()
					return nil, err
				}
			}
			return n, nil
		}))

		return nil
	}
}
