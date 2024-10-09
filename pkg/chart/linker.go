package chart

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// LinkerConfig holds the configuration for the linker, including the scheme and hooks for loading/unloading symbols.
type LinkerConfig struct {
	Scheme      *scheme.Scheme      // Specifies the scheme, which defines symbol and node behavior.
	LoadHooks   []symbol.LoadHook   // A list of hooks to be executed when symbols are loaded.
	UnloadHooks []symbol.UnloadHook // A list of hooks to be executed when symbols are unloaded.
}

// Linker manages chart loading and unloading.
type Linker struct {
	scheme      *scheme.Scheme
	codecs      map[string]scheme.Codec
	loadHooks   []symbol.LoadHook
	unloadHooks []symbol.UnloadHook
	mu          sync.RWMutex
}

var _ LinkHook = (*Linker)(nil)
var _ UnlinkHook = (*Linker)(nil)

// NewLinker creates a new Linker.
func NewLinker(config LinkerConfig) *Linker {
	return &Linker{
		scheme:      config.Scheme,
		codecs:      make(map[string]scheme.Codec),
		loadHooks:   config.LoadHooks,
		unloadHooks: config.UnloadHooks,
	}
}

// Link loads the chart, creating nodes and symbols.
func (l *Linker) Link(chrt *Chart) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	kind := chrt.GetName()
	codec := l.codecs[kind]

	if l.scheme.Codec(kind) != codec {
		return nil
	}

	codec = scheme.CodecFunc(func(sp spec.Spec) (node.Node, error) {
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

			symbols = append(symbols, &symbol.Symbol{
				Spec: sp,
				Node: n,
			})
		}

		table := symbol.NewTable(symbol.TableOption{
			LoadHooks:   l.loadHooks,
			UnloadHooks: l.unloadHooks,
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
					if port.Name == "" || sb.ID() == port.ID || sb.Name() == port.Name {
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

	l.scheme.AddKnownType(kind, &spec.Unstructured{})
	l.scheme.AddCodec(kind, codec)
	l.codecs[kind] = codec
	return nil
}

// Unlink removes the chart from the scheme.
func (l *Linker) Unlink(chrt *Chart) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	kind := chrt.GetName()
	codec := l.codecs[kind]

	if l.scheme.Codec(kind) != codec {
		return nil
	}

	l.scheme.RemoveKnownType(kind)
	l.scheme.RemoveCodec(kind)
	delete(l.codecs, kind)
	return nil
}
