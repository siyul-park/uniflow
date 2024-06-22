package language

import (
	"encoding/json"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/expr-lang/expr"
	"github.com/siyul-park/uniflow/plugin/internal/js"
	"github.com/xiatechs/jsonata-go"
	"gopkg.in/yaml.v3"
)

const (
	Expr       = "expr"
	Text       = "text"
	Typescript = "typescript"
	Javascript = "javascript"
	JSON       = "json"
	JSONata    = "jsonata"
	YAML       = "yaml"
)

var exprSymbols = []string{"#", ".", "[", "]", "&", "|", "-", "+", "*", "/", "%", "=", ">", "<"}
var jsonataSymbols = []string{"$", ".", "[", "]", "&", "|", "-", "+", "*", "/", "%", "=", ">", "<"}
var javascriptSymbols = []string{".", "[", "]", "&", "|", "-", "+", "*", "/", "%", "=", ">", "<"}

func Detect(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return Text
	}

	var raw any
	if err := json.Unmarshal([]byte(value), &raw); err == nil {
		return JSON
	}
	if _, err := expr.Compile(value); err == nil {
		for _, symbol := range exprSymbols {
			if strings.Contains(value, symbol) {
				return Expr
			}
		}
	}
	if _, err := jsonata.Compile(value); err == nil {
		for _, symbol := range jsonataSymbols {
			if strings.Contains(value, symbol) {
				return JSONata
			}
		}
	}
	if _, err := js.Transform(value, api.TransformOptions{Loader: api.LoaderJS}); err == nil {
		for _, symbol := range javascriptSymbols {
			if strings.Contains(value, symbol) {
				return Javascript
			}
		}
	}
	if _, err := js.Transform(value, api.TransformOptions{Loader: api.LoaderTS}); err == nil {
		for _, symbol := range javascriptSymbols {
			if strings.Contains(value, symbol) {
				return Typescript
			}
		}
	}
	if err := yaml.Unmarshal([]byte(value), &raw); err == nil {
		if _, ok := raw.(string); !ok || value != raw {
			return YAML
		}
	}
	return Text
}
