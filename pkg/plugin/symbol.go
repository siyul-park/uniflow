package plugin

import "reflect"

// Symbols maps package paths to their exported symbols.
var Symbols = map[string]map[string]reflect.Value{}

// MapTypes maps functions to types they handle specially.
var MapTypes = map[reflect.Value][]reflect.Type{}

func init() {
	Symbols["."] = map[string]reflect.Value{
		"MapTypes": reflect.ValueOf(MapTypes),
	}
}

//go:generate go install github.com/traefik/yaegi/cmd/yaegi@latest

//go:generate yaegi extract github.com/siyul-park/uniflow/pkg/driver
//go:generate yaegi extract github.com/siyul-park/uniflow/pkg/hook
//go:generate yaegi extract github.com/siyul-park/uniflow/pkg/language
//go:generate yaegi extract github.com/siyul-park/uniflow/pkg/meta
//go:generate yaegi extract github.com/siyul-park/uniflow/pkg/node
//go:generate yaegi extract github.com/siyul-park/uniflow/pkg/packet
//_go:generate yaegi extract github.com/siyul-park/uniflow/pkg/plugin
//go:generate yaegi extract github.com/siyul-park/uniflow/pkg/port
//go:generate yaegi extract github.com/siyul-park/uniflow/pkg/process
//go:generate yaegi extract github.com/siyul-park/uniflow/pkg/runtime
//go:generate yaegi extract github.com/siyul-park/uniflow/pkg/scheme
//go:generate yaegi extract github.com/siyul-park/uniflow/pkg/spec
//go:generate yaegi extract github.com/siyul-park/uniflow/pkg/symbol
//go:generate yaegi extract github.com/siyul-park/uniflow/pkg/testing
//go:generate yaegi extract github.com/siyul-park/uniflow/pkg/types
//go:generate yaegi extract github.com/siyul-park/uniflow/pkg/value

//go:generate yaegi extract github.com/gofrs/uuid
