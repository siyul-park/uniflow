package control

import "github.com/pkg/errors"

const (
	LangText       = "text"
	LangTypescript = "typescript"
	LangJSON       = "json"
	LangYAML       = "yaml"
	LangJavascript = "javascript"
	LangJSONata    = "jsonata"
)

var ErrUnsupportedLanguage = errors.New("language not supported")
