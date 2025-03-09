package testing

import (
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"sync"
)

// Mocker manages the creation and removal of mock connections between symbols based on port matching.
type Mocker struct {
	symbols    map[uuid.UUID]*symbol.Symbol
	namespaces map[string]map[string]uuid.UUID
	mu         sync.RWMutex
}

// NewMocker initializes and returns a new Mocker instance.
func NewMocker() *Mocker {
	return &Mocker{
		symbols:    make(map[uuid.UUID]*symbol.Symbol),
		namespaces: make(map[string]map[string]uuid.UUID),
	}
}

var _ symbol.LoadHook = (*Mocker)(nil)
var _ symbol.UnloadHook = (*Mocker)(nil)

// Mock establishes mock connections between the provided symbol and other symbols based on matching ports.
func (m *Mocker) Mock(proc *process.Process, mock *symbol.Symbol) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sb, ok := m.symbols[mock.ID()]
	if !ok || sb == mock {
		return nil
	}

	for name, ports := range mock.Ports() {
		out := mock.Out(name)
		for _, port := range ports {
			id := port.ID
			if id == uuid.Nil {
				id = m.lookup(sb.Namespace(), port.Name)
			}

			if ref, ok := m.symbols[id]; ok {
				if ref.Namespace() == sb.Namespace() {
					in := ref.In(port.Port)
					if out != nil && in != nil {
						out.Link(in)
					}
				}
			}
		}
	}

	for _, ref := range m.symbols {
		if ref.Namespace() != sb.Namespace() {
			continue
		}
		for name, ports := range ref.Ports() {
			out := ref.Out(name)
			for _, port := range ports {
				id := port.ID
				if id == uuid.Nil {
					id = m.lookup(ref.Namespace(), port.Name)
				}

				if id == mock.ID() {
					writer := out.Open(proc)
					if in := mock.In(port.Port); in != nil {
						reader := in.Open(proc)
						writer.Link(reader)
					}
					if in := sb.In(port.Port); in != nil {
						reader := in.Open(proc)
						writer.Unlink(reader)
					}
				}
			}
		}
	}
	return nil
}

// Load adds the symbol to the symbol map and its namespace.
func (m *Mocker) Load(sb *symbol.Symbol) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.symbols[sb.ID()] = sb

	if sb.Name() != "" {
		ns, ok := m.namespaces[sb.Namespace()]
		if !ok {
			ns = make(map[string]uuid.UUID)
			m.namespaces[sb.Namespace()] = ns
		}
		ns[sb.Name()] = sb.ID()
	}
	return nil
}

// Unload removes the symbol from the symbol map and cleans up its namespace.
func (m *Mocker) Unload(sb *symbol.Symbol) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clean up the namespace if the symbol has a name.
	if sb.Name() != "" {
		if ns, ok := m.namespaces[sb.Namespace()]; ok {
			delete(ns, sb.Name())
			if len(ns) == 0 {
				delete(m.namespaces, sb.Namespace())
			}
		}
	}

	delete(m.symbols, sb.ID())
	return nil
}

func (m *Mocker) lookup(namespace, name string) uuid.UUID {
	if ns, ok := m.namespaces[namespace]; ok {
		return ns[name]
	}
	return uuid.Nil
}
