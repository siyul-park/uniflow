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
		if !strings.Contains(value, ".") && !strings.Contains(value, "$") {
			return Text
		}
		return JSONata
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
