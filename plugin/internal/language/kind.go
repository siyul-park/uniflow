package language

import (
	"encoding/json"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/siyul-park/uniflow/plugin/internal/js"
	"github.com/xiatechs/jsonata-go"
	"gopkg.in/yaml.v3"
	"strings"
)

const (
	Text       = "text"
	Typescript = "typescript"
	Javascript = "javascript"
	JSON       = "json"
	JSONata    = "jsonata"
	YAML       = "yaml"
)

var symbols = []string{"$", ".", "[", "]", "&", "|", "-", "+", "*", "/", "%", "=", ">", "<"}

func Detect(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return Text
	}

	var raw any
	if err := json.Unmarshal([]byte(value), &raw); err == nil {
		return JSON
	}
	if _, err := jsonata.Compile(value); err == nil {
		for _, symbol := range symbols {
			if strings.Contains(value, symbol) {
				return JSONata
			}
		}
		return Text
	}
	if _, err := js.Transform(value, api.TransformOptions{Loader: api.LoaderJS}); err == nil {
		return Javascript
	}
	if _, err := js.Transform(value, api.TransformOptions{Loader: api.LoaderTS}); err == nil {
		return Typescript
	}
	if _, err := js.Transform(value, api.TransformOptions{Loader: api.LoaderTS}); err == nil {
		return Typescript
	}
	if err := yaml.Unmarshal([]byte(value), &raw); err == nil {
		if _, ok := raw.(string); ok && value == raw {
			return Text
		}
		return YAML
	}
	return Text
}
