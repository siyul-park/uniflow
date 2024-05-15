package hook

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Hook represents a collection of hook functions.
type Hook struct {
	loadHooks   []symbol.LoadHook
	unloadHooks []symbol.UnloadHook
	mu          sync.RWMutex
}

var _ symbol.LoadHook = (*Hook)(nil)
var _ symbol.UnloadHook = (*Hook)(nil)

// New creates a new instance of Hook.
func New() *Hook {
	return &Hook{}
}

// AddLoadHook adds a LoadHook function to the Hook.
func (h *Hook) AddLoadHook(hook symbol.LoadHook) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.loadHooks = append(h.loadHooks, hook)
}

// AddUnloadHook adds an UnloadHook function to the Hook.
func (h *Hook) AddUnloadHook(hook symbol.UnloadHook) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.unloadHooks = append(h.unloadHooks, hook)
}

// Load executes LoadHooks on the provided node.
func (h *Hook) Load(sym *symbol.Symbol) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.loadHooks {
		if err := hook.Load(sym); err != nil {
			return err
		}
	}
	return nil
}

// Unload executes UnloadHooks on the provided node.
func (h *Hook) Unload(sym *symbol.Symbol) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for i := len(h.unloadHooks); i >= 0; i-- {
		hook := h.unloadHooks[i]
		if err := hook.Unload(sym); err != nil {
			return err
		}
	}
	return nil
}
