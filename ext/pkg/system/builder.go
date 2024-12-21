package system

import (
	"context"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/siyul-park/uniflow/pkg/types"
	"reflect"
)

type SchemeRegister struct {
	syscalls map[string]func(context.Context, []any) ([]any, error)
	signals  map[string]func(context.Context) (<-chan any, error)
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
func AddToScheme() *SchemeRegister {
	return &SchemeRegister{
		syscalls: make(map[string]func(context.Context, []any) ([]any, error)),
		signals:  make(map[string]func(context.Context) (<-chan any, error)),
	}
}

func (r *SchemeRegister) AddToScheme(s *scheme.Scheme) error {
	definitions := []struct {
		kind  string
		codec scheme.Codec
		spec  spec.Spec
	}{
		{KindSyscall, NewSyscallNodeCodec(r.syscalls), &SyscallNodeSpec{}},
		{KindSignal, NewSignalNodeCodec(r.signals), &SignalNodeSpec{}},
	}

	for _, def := range definitions {
		s.AddKnownType(def.kind, def.spec)
		s.AddCodec(def.kind, def.codec)
	}

	return nil
}

func (r *SchemeRegister) SetSignal(topic string, fn any) error {
	var signal func(context.Context) (<-chan any, error)
	if s, ok := fn.(func(context.Context) (<-chan any, error)); ok {
		signal = s
	} else if s, ok := fn.(func(context.Context) <-chan any); ok {
		signal = func(ctx context.Context) (<-chan any, error) {
			return s(ctx), nil
		}
	} else if s, ok := fn.(func() (<-chan any, error)); ok {
		signal = func(_ context.Context) (<-chan any, error) {
			return s()
		}
	} else if s, ok := fn.(func() <-chan any); ok {
		signal = func(_ context.Context) (<-chan any, error) {
			return s(), nil
		}
	}
	if signal == nil {
		return errors.WithStack(encoding.ErrUnsupportedType)
	}

	r.signals[topic] = signal
	return nil
}

func (r *SchemeRegister) Signal(topic string) func(context.Context) (<-chan any, error) {
	return r.signals[topic]
}

func (r *SchemeRegister) SetSyscall(opcode string, fn any) error {
	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		return errors.WithStack(encoding.ErrUnsupportedType)
	}

	typeContext := reflect.TypeOf((*context.Context)(nil)).Elem()
	typeError := reflect.TypeOf((*error)(nil)).Elem()

	fnType := fnValue.Type()
	numIn := fnType.NumIn()
	numOut := fnType.NumOut()

	r.syscalls[opcode] = func(ctx context.Context, arguments []any) ([]any, error) {
		ins := make([]reflect.Value, numIn)
		offset := 0

		if numIn > 0 && fnType.In(0).Implements(typeContext) {
			ins[0] = reflect.ValueOf(ctx)
			offset++
		}

		for i := offset; i < numIn; i++ {
			if i-offset < len(arguments) {
				arg, err := types.Marshal(arguments[i-offset])
				if err != nil {
					return nil, err
				}
				in := reflect.New(fnType.In(i)).Interface()
				if err := types.Unmarshal(arg, in); err != nil {
					return nil, err
				}
				ins[i] = reflect.ValueOf(in).Elem()
			} else {
				ins[i] = reflect.Zero(fnType.In(i))
			}
		}

		outs := fnValue.Call(ins)

		if numOut > 0 && fnType.Out(numOut-1).Implements(typeError) {
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
	return nil
}

func (r *SchemeRegister) Syscall(opcode string) func(context.Context, []any) ([]any, error) {
	return r.syscalls[opcode]
}
