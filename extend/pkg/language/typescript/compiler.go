package typescript

import (
	"github.com/evanw/esbuild/pkg/api"
	"github.com/siyul-park/uniflow/extend/pkg/language"
	"github.com/siyul-park/uniflow/extend/pkg/language/javascript"
)

const Language = "typescript"

func NewCompiler() language.Compiler {
	return javascript.NewCompiler(api.TransformOptions{
		Format: api.FormatCommonJS,
		Loader: api.LoaderTS,
	})
}
