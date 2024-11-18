package io

import (
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// AddToScheme returns a function that adds node types and codecs to the provided spec.
func AddToScheme(fs FileSystem) scheme.Register {
	return scheme.RegisterFunc(func(s *scheme.Scheme) error {
		definitions := []struct {
			kind  string
			codec scheme.Codec
			spec  spec.Spec
		}{
			{KindSQL, NewSQLNodeCodec(), &SQLNodeSpec{}},
			{KindPrint, NewPrintNodeCodec(fs), &PrintNodeSpec{}},
			{KindScan, NewScanNodeCodec(fs), &ScanNodeSpec{}},
		}

		for _, def := range definitions {
			s.AddKnownType(def.kind, def.spec)
			s.AddCodec(def.kind, def.codec)
		}

		return nil
	})

}
