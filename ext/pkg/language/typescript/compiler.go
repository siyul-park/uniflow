package typescript

import (
	"github.com/evanw/esbuild/pkg/api"
	"github.com/siyul-park/uniflow/ext/pkg/language"
	"github.com/siyul-park/uniflow/ext/pkg/language/javascript"
)

const Language = "typescript"

func NewCompiler() language.Compiler {
	return javascript.NewCompiler(api.TransformOptions{
		Format: api.FormatCommonJS,
		Loader: api.LoaderTS,
		Target: api.ES2016,
	})
}
