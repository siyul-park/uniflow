package chart

import (
	"slices"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/siyul-park/uniflow/pkg/template"
	"github.com/siyul-park/uniflow/pkg/types"
)

// LinkerConfig holds the hook and scheme configuration.
type LinkerConfig struct {
	Hook   *hook.Hook     // Manages symbol lifecycle events.
	Scheme *scheme.Scheme // Defines symbol and node behavior.
}

// Linker manages chart loading and unloading.
type Linker struct {
	hook   *hook.Hook
	scheme *scheme.Scheme
}

var _ LoadHook = (*Linker)(nil)
var _ UnloadHook = (*Linker)(nil)

// NewLinker creates a new Linker.
func NewLinker(config LinkerConfig) *Linker {
	return &Linker{
		hook:   config.Hook,
		scheme: config.Scheme,
	}
}

// Load loads the chart, creating nodes and symbols.
func (l *Linker) Load(chrt *Chart) error {
	if slices.Contains(l.scheme.Kinds(), chrt.GetName()) {
		return nil
	}

	codec := scheme.CodecFunc(func(sp spec.Spec) (node.Node, error) {
		doc, err := types.Marshal(sp)
		if err != nil {
			return nil, err
		}

		env := map[string][]spec.Value{}
		for key, vals := range chrt.GetEnv() {
			for _, val := range vals {
				if val.ID == uuid.Nil && val.Name == "" {
					v, err := template.Execute(val.Value, types.InterfaceOf(doc))
					if err != nil {
						return nil, err
					}
					val.Value = v
				}
				env[key] = append(env[key], spec.Value{Value: val.Value})
			}
		}

		symbols := make([]*symbol.Symbol, 0, len(chrt.GetSpecs()))
		for _, sp := range chrt.GetSpecs() {
			doc, err := types.Marshal(sp)
			if err != nil {
				return nil, err
			}

			unstructured := &spec.Unstructured{}
			if err := types.Unmarshal(doc, unstructured); err != nil {
				return nil, err
			}

			unstructured.SetEnv(env)

			bind, err := spec.Bind(unstructured)
			if err != nil {
				return nil, err
			}

			decode, err := l.scheme.Decode(bind)
			if err != nil {
				return nil, err
			}

			n, err := l.scheme.Compile(decode)
			if err != nil {
				for _, sb := range symbols {
					sb.Close()
				}
				return nil, err
			}

			symbols = append(symbols, &symbol.Symbol{Spec: decode, Node: n})
		}

		var loadHooks []symbol.LoadHook
		var unloadHook []symbol.UnloadHook
		if l.hook != nil {
			loadHooks = append(loadHooks, l.hook)
			unloadHook = append(unloadHook, l.hook)
		}

		table := symbol.NewTable(symbol.TableOption{
			LoadHooks:   loadHooks,
			UnloadHooks: unloadHook,
		})

		for _, sb := range symbols {
			if err := table.Insert(sb); err != nil {
				table.Close()
				for _, sb := range symbols {
					sb.Close()
				}
				return nil, err
			}
		}

		return nil, nil
	})

	l.scheme.AddKnownType(chrt.GetName(), &spec.Unstructured{})
	l.scheme.AddCodec(chrt.GetName(), codec)

	return nil
}

// Unload removes the chart from the scheme.
func (l *Linker) Unload(chrt *Chart) error {
	l.scheme.RemoveKnownType(chrt.GetName())
	l.scheme.RemoveCodec(chrt.GetName())

	return nil
}
