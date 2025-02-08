package control

import (
	"github.com/siyul-park/uniflow/ext/pkg/language"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
)

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
			{KindCache, NewCacheNodeCodec(), &CacheNodeSpec{}},
			{KindFor, NewForNodeCodec(), &ForNodeSpec{}},
			{KindFork, NewForkNodeCodec(), &ForkNodeSpec{}},
			{KindIf, NewIfNodeCodec(expr), &IfNodeSpec{}},
			{KindMerge, NewMergeNodeCodec(), &MergeNodeSpec{}},
			{KindNOP, NewNOPNodeCodec(), &NOPNodeSpec{}},
			{KindPipe, NewPipeNodeCodec(), &PipeNodeSpec{}},
			{KindRetry, NewRetryNodeCodec(), &RetryNodeSpec{}},
			{KindSession, NewSessionNodeCodec(), &SessionNodeSpec{}},
			{KindSleep, NewSleepNodeCodec(), &SleepNodeSpec{}},
			{KindSnippet, NewSnippetNodeCodec(module), &SnippetNodeSpec{}},
			{KindSplit, NewSplitNodeCodec(), &SplitNodeSpec{}},
			{KindStep, NewStepNodeCodec(s), &StepNodeSpec{}},
			{KindSwitch, NewSwitchNodeCodec(expr), &SwitchNodeSpec{}},
			{KindThrow, NewThrowNodeCodec(), &ThrowNodeSpec{}},
			{KindTry, NewTryNodeCodec(), &TryNodeSpec{}},
			{KindAssert, NewAssertNodeCodec(expr), &AssertNodeSpec{}},
		}

		for _, def := range definitions {
			s.AddKnownType(def.kind, def.spec)
			s.AddCodec(def.kind, def.codec)
		}

		return nil
	})
}
