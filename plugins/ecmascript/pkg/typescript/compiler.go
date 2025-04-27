package typescript

import (
	"github.com/evanw/esbuild/pkg/api"
	"github.com/siyul-park/uniflow/pkg/language"

	"github.com/siyul-park/uniflow/plugins/ecmascript/pkg/javascript"
)

const Language = "typescript"

func NewCompiler(options ...api.TransformOptions) language.Compiler {
	return javascript.NewCompiler(append(
		[]api.TransformOptions{{
			Format: api.FormatCommonJS,
			Loader: api.LoaderTS,
			Target: api.ES2016,
		}},
		options...,
	)...)
}
