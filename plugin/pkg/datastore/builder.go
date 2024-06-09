package datastore

import (
	"github.com/siyul-park/uniflow/pkg/spec"
)

// AddToScheme returns a function that adds node types and codecs to the provided spec.
func AddToScheme() func(*spec.Scheme) error {
	return func(s *spec.Scheme) error {
		s.AddKnownType(KindRDB, &RDBNodeSpec{})
		s.AddCodec(KindRDB, NewRDBNodeCodec())

		s.AddKnownType(KindWrite, &WriteNodeSpec{})
		s.AddCodec(KindWrite, NewWriteNodeCodec())

		return nil
	}
}
