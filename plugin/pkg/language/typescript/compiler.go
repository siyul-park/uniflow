package typescript

import (
	"github.com/evanw/esbuild/pkg/api"
	"github.com/siyul-park/uniflow/plugin/pkg/language"
	"github.com/siyul-park/uniflow/plugin/pkg/language/javascript"
)

func NewCompiler() language.Compiler {
	return javascript.NewCompiler(api.TransformOptions{
		Format: api.FormatCommonJS,
		Loader: api.LoaderTS,
	})
}
