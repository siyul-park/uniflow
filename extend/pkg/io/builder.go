package io

import "github.com/siyul-park/uniflow/pkg/scheme"

// AddToScheme returns a function that adds node types and codecs to the provided spec.
func AddToScheme() func(*scheme.Scheme) error {
	return func(s *scheme.Scheme) error {
		s.AddKnownType(KindRDB, &RDBNodeSpec{})
		s.AddCodec(KindRDB, NewRDBNodeCodec())

		s.AddKnownType(KindRead, &ReadNodeSpec{})
		s.AddCodec(KindRead, NewReadNodeCodec())

		s.AddKnownType(KindWrite, &WriteNodeSpec{})
		s.AddCodec(KindWrite, NewWriteNodeCodec())

		return nil
	}
}
