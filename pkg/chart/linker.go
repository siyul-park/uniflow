package chart

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Linker manages chart loading and unloading.
type Linker struct {
	scheme *scheme.Scheme
	codecs map[string]scheme.Codec
	mu     sync.RWMutex
}

var _ LinkHook = (*Linker)(nil)
var _ UnlinkHook = (*Linker)(nil)

// NewLinker creates a new Linker.
func NewLinker(s *scheme.Scheme) *Linker {
	return &Linker{
		scheme: s,
		codecs: make(map[string]scheme.Codec),
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

		n := symbol.NewCluster(symbols)

		for name, ports := range chrt.GetInbound() {
			for _, port := range ports {
				n.Inbound(name, port)
			}
		}

		for name, ports := range chrt.GetOutbound() {
			for _, port := range ports {
				n.Outbound(name, port)
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
