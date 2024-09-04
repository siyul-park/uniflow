package io

import "github.com/siyul-park/uniflow/pkg/scheme"

// AddToScheme returns a function that adds node types and codecs to the provided spec.
func AddToScheme(fs FileSystem) scheme.Register {
	return scheme.RegisterFunc(func(s *scheme.Scheme) error {
		s.AddKnownType(KindRDB, &RDBNodeSpec{})
		s.AddCodec(KindRDB, NewRDBNodeCodec())

		s.AddKnownType(KindRead, &ReadNodeSpec{})
		s.AddCodec(KindRead, NewReadNodeCodec(fs))

		s.AddKnownType(KindWrite, &WriteNodeSpec{})
		s.AddCodec(KindWrite, NewWriteNodeCodec(fs))

		return nil
	})
}
