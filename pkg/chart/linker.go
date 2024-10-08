package chart

import (
	"slices"

	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
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
		specs, err := chrt.Build(sp)
		if err != nil {
			return nil, err
		}

		symbols := make([]*symbol.Symbol, 0, len(specs))
		for _, sp := range specs {
			n, err := l.scheme.Compile(sp)
			if err != nil {
				for _, sb := range symbols {
					sb.Close()
				}
				return nil, err
			}
			symbols = append(symbols, &symbol.Symbol{Spec: sp, Node: n})
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

		n := NewClusterNode(table)
		for name, ports := range chrt.GetPorts() {
			for _, port := range ports {
				for _, sb := range symbols {
					if (sb.ID() == port.ID) || (sb.Name() != "" && sb.Name() == port.Name) {
						if in := sb.In(port.Port); in != nil {
							n.Inbound(name, in)
						}
						if out := sb.Out(port.Port); out != nil {
							n.Outbound(name, out)
						}
					}
				}
			}
		}
		return n, nil
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
