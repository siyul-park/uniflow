package system

import (
	"context"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/siyul-park/uniflow/pkg/types"
	"reflect"
)

type SchemeRegister struct {
	operators map[string]any
}

var _ scheme.Register = (*SchemeRegister)(nil)

// AddToHook returns a function that adds hook to the provided hook.
func AddToHook() hook.Register {
	return hook.RegisterFunc(func(h *hook.Hook) error {
		h.AddLoadHook(symbol.LoadFunc(func(sb *symbol.Symbol) error {
			if n, ok := node.Unwrap(sb).(*SignalNode); ok {
				n.Listen()
			}
			return nil
		}))
		h.AddUnloadHook(symbol.UnloadFunc(func(sb *symbol.Symbol) error {
			if n, ok := node.Unwrap(sb).(*SignalNode); ok {
				n.Shutdown()
			}
			return nil
		}))
		return nil
	})
}

// AddToScheme returns a function that adds node types and codecs to the provided spec.
func AddToScheme(operators map[string]any) *SchemeRegister {
	return &SchemeRegister{operators: operators}
}

func (r *SchemeRegister) AddToScheme(s *scheme.Scheme) error {
	functions := make(map[string]func(context.Context, []any) ([]any, error))
	signals := make(map[string]func(context.Context) (<-chan any, error))

	for opcode := range r.operators {
		if signal := r.Signal(opcode); signal != nil {
			signals[opcode] = signal
		} else if fn := r.Function(opcode); fn != nil {
			functions[opcode] = fn
		}
	}

	definitions := []struct {
		kind  string
		codec scheme.Codec
		spec  spec.Spec
	}{
		{KindNative, NewNativeNodeCodec(functions), &NativeNodeSpec{}},
		{KindSignal, NewSignalNodeCodec(signals), &SignalNodeSpec{}},
	}

	for _, def := range definitions {
		s.AddKnownType(def.kind, def.spec)
		s.AddCodec(def.kind, def.codec)
	}

	return nil
}

func (r *SchemeRegister) Signal(opcode string) func(context.Context) (<-chan any, error) {
	op, ok := r.operators[opcode]
	if !ok {
		return nil
	}

	if signal, ok := op.(func(context.Context) (<-chan any, error)); ok {
		return signal
	} else if signal, ok := op.(func(context.Context) <-chan any); ok {
		return func(ctx context.Context) (<-chan any, error) {
			return signal(ctx), nil
		}
	} else if signal, ok := op.(func() (<-chan any, error)); ok {
		return func(_ context.Context) (<-chan any, error) {
			return signal()
		}
	} else if signal, ok := op.(func() <-chan any); ok {
		return func(_ context.Context) (<-chan any, error) {
			return signal(), nil
		}
	} else {
		return nil
	}
}

func (r *SchemeRegister) Function(opcode string) func(context.Context, []any) ([]any, error) {
	op, ok := r.operators[opcode]
	if !ok {
		return nil
	}

	fn := reflect.ValueOf(op)
	if fn.Kind() != reflect.Func {
		return nil
	}

	typeContext := reflect.TypeOf((*context.Context)(nil)).Elem()
	typeError := reflect.TypeOf((*error)(nil)).Elem()

	opType := fn.Type()
	numIn := opType.NumIn()
	numOut := opType.NumOut()

	return func(ctx context.Context, arguments []any) ([]any, error) {
		ins := make([]reflect.Value, numIn)
		offset := 0

		if numIn > 0 && opType.In(0).Implements(typeContext) {
			ins[0] = reflect.ValueOf(ctx)
			offset++
		}

		for i := offset; i < numIn; i++ {
			if i-offset < len(arguments) {
				arg, err := types.Marshal(arguments[i-offset])
				if err != nil {
					return nil, err
				}
				in := reflect.New(opType.In(i)).Interface()
				if err := types.Unmarshal(arg, in); err != nil {
					return nil, err
				}
				ins[i] = reflect.ValueOf(in).Elem()
			} else {
				ins[i] = reflect.Zero(opType.In(i))
			}
		}

		outs := fn.Call(ins)

		if numOut > 0 && opType.Out(numOut-1).Implements(typeError) {
			if err, ok := outs[numOut-1].Interface().(error); ok && err != nil {
				return nil, err
			}
			outs = outs[:numOut-1]
		}

		returns := make([]any, len(outs))
		for i, out := range outs {
			returns[i] = out.Interface()
		}
		return returns, nil
	}
}
