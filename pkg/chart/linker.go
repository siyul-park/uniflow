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
	loadHooks   symbol.LoadHooks
	unloadHooks symbol.UnloadHooks
	mu          sync.RWMutex
}

var _ LinkHook = (*Linker)(nil)
var _ UnlinkHook = (*Linker)(nil)
var _ symbol.LoadHook = (*Linker)(nil)
var _ symbol.UnloadHook = (*Linker)(nil)

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
			unstructured := &spec.Unstructured{}
			if err := spec.Convert(sp, unstructured); err != nil {
				for _, sb := range symbols {
					_ = sb.Close()
				}
				return nil, err
			} else if err := unstructured.Build(); err != nil {
				for _, sb := range symbols {
					_ = sb.Close()
				}
				return nil, err
			} else if decode, err := l.scheme.Decode(unstructured); err != nil {
				for _, sb := range symbols {
					_ = sb.Close()
				}
				return nil, err
			} else {
				sp = decode
			}

			n, err := l.scheme.Compile(sp)
			if err != nil {
				for _, sb := range symbols {
					_ = sb.Close()
				}
				return nil, err
			}

			symbols = append(symbols, &symbol.Symbol{
				Spec: sp,
				Node: n,
			})
		}

		n := NewClusterNode(symbols, symbol.TableOption{
			LoadHooks:   []symbol.LoadHook{l.loadHooks},
			UnloadHooks: []symbol.UnloadHook{l.unloadHooks},
		})

		for name, ports := range chrt.GetInbound() {
			for _, port := range ports {
				for _, sb := range symbols {
					if port.Name == "" || sb.ID() == port.ID || sb.Name() == port.Name {
						n.Inbound(name, sb.ID(), port.Port)
					}
				}
			}
		}

		for name, ports := range chrt.GetOutbound() {
			for _, port := range ports {
				for _, sb := range symbols {
					if port.Name == "" || sb.ID() == port.ID || sb.Name() == port.Name {
						n.Outbound(name, sb.ID(), port.Port)
					}
				}
			}
		}

		return n, nil
	})

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

	l.scheme.RemoveCodec(kind)
	delete(l.codecs, kind)
	return nil
}

// Load loads the symbol's node if it is a ClusterNode.
func (l *Linker) Load(sb *symbol.Symbol) error {
	n := sb.Node
	if n, ok := n.(*ClusterNode); ok {
		if err := n.Load(nil); err != nil {
			return err
		}
	}
	return nil
}

// Unload unloads the symbol's node if it is a ClusterNode.
func (l *Linker) Unload(sb *symbol.Symbol) error {
	n := sb.Node
	if n, ok := n.(*ClusterNode); ok {
		if err := n.Unload(nil); err != nil {
			return err
		}
	}
	return nil
}
