package hook

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/event"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Hook represents a collection of hook functions.
type Hook struct {
	loadHooks   []LoadHook
	unloadHooks []UnloadHook
	mu          sync.RWMutex
}

var _ LoadHook = (*Hook)(nil)
var _ UnloadHook = (*Hook)(nil)

// New creates a new instance of Hook.
func New() *Hook {
	return &Hook{}
}

// Subscribe subscribes to the event broker and asynchronously processes load and unload events.
func (h *Hook) Subscribe(broker *event.Broker) {
	load := broker.Consumer(symbol.TopicLoad)
	unload := broker.Consumer(symbol.TopicUnload)

	go func() {
		for e := range load.Consume() {
			if n, ok := e.Data().(node.Node); ok {
				_ = h.Load(n)
			}
			e.Wait(-1)
		}
	}()

	go func() {
		for e := range unload.Consume() {
			if n, ok := e.Data().(node.Node); ok {
				_ = h.Unload(n)
			}
			e.Wait(-1)
		}
	}()
}

// AddLoadHook adds a LoadHook function to the Hook.
func (h *Hook) AddLoadHook(hook LoadHook) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.loadHooks = append(h.loadHooks, hook)
}

// AddUnloadHook adds an UnloadHook function to the Hook.
func (h *Hook) AddUnloadHook(hook UnloadHook) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.unloadHooks = append(h.unloadHooks, hook)
}

// Load executes LoadHooks on the provided node.
func (h *Hook) Load(n node.Node) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.loadHooks {
		if err := hook.Load(n); err != nil {
			return err
		}
	}
	return nil
}

// Unload executes UnloadHooks on the provided node.
func (h *Hook) Unload(n node.Node) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.unloadHooks {
		if err := hook.Unload(n); err != nil {
			return err
		}
	}
	return nil
}
