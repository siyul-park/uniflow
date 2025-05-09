package plugin

import (
	"context"
	"encoding/json"
	"plugin"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

// Plugin defines the interface that dynamic plugins must implement.
type Plugin interface {
	Load(ctx context.Context) error
	Unload(ctx context.Context) error
}

var (
	ErrInvalidSignature  = errors.New("invalid signature")
	ErrMissingDependency = errors.New("missing dependency")
)

// Open loads a plugin from the given path and returns an instance created by the plugin's New function.
// The manifest is marshaled to JSON and passed as input to New.
func Open(path string, manifest any) (Plugin, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	p, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}

	sym, err := p.Lookup("New")
	if err != nil {
		return nil, err
	}

	val := reflect.ValueOf(sym)
	typ := reflect.TypeOf(sym)

	var ins []reflect.Value
	for i := 0; i < typ.NumIn(); i++ {
		data, err := json.Marshal(manifest)
		if err != nil {
			return nil, err
		}
		in := reflect.New(val.Type().In(i))
		if err := json.Unmarshal(data, in.Interface()); err != nil {
			return nil, err
		}
		if err := validate.Struct(in.Interface()); err != nil {
			return nil, err
		}
		ins = append(ins, in.Elem())
	}

	outs := val.Call(ins)
	if len(outs) == 0 {
		return nil, errors.WithStack(ErrInvalidSignature)
	}

	if len(outs) > 1 {
		if err, ok := outs[len(outs)-1].Interface().(error); ok && err != nil {
			return nil, err
		}
	}

	v, ok := outs[0].Interface().(Plugin)
	if !ok {
		return nil, errors.WithStack(ErrInvalidSignature)
	}
	return v, nil
}
