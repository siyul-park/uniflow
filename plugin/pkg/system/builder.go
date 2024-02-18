package system

import (
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// AddToScheme returns a function that adds node types and codecs to the provided scheme.
func AddToScheme(table *SyscallTable) func(*scheme.Scheme) error {
	if table == nil {
		table = NewSyscallTable()
	}

	return func(s *scheme.Scheme) error {
		s.AddKnownType(KindSyscall, &SyscallNodeSpec{})
		s.AddCodec(KindSyscall, NewSyscallNodeCodec(table))

		return nil
	}
}
