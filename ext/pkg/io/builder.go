package io

import "github.com/siyul-park/uniflow/pkg/scheme"

// AddToScheme returns a function that adds node types and codecs to the provided spec.
func AddToScheme(fs FileSystem) scheme.Register {
	return scheme.RegisterFunc(func(s *scheme.Scheme) error {
		s.AddKnownType(KindSQL, &SQLNodeSpec{})
		s.AddCodec(KindSQL, NewSQLNodeCodec())

		s.AddKnownType(KindPrint, &PrintNodeSpec{})
		s.AddCodec(KindPrint, NewPrintNodeCodec(fs))

		s.AddKnownType(KindScan, &ScanNodeSpec{})
		s.AddCodec(KindScan, NewScanNodeCodec(fs))

		return nil
	})
}
