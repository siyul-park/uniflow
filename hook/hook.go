package hook

import (
	"sync"

	"github.com/siyul-park/uniflow/symbol"
)

// Hook represents a collection of hook functions that can be executed on symbols.
type Hook struct {
	loadHooks   symbol.LoadHooks
	unloadHooks symbol.UnloadHooks
	mu          sync.RWMutex
}

var _ symbol.LoadHook = (*Hook)(nil)
var _ symbol.UnloadHook = (*Hook)(nil)

// New creates a new instance of Hook.
func New() *Hook {
	return &Hook{}
}

// AddLoadHook adds a LoadHook function to the Hook.
func (h *Hook) AddLoadHook(hook symbol.LoadHook) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, h := range h.loadHooks {
		if h == hook {
			return false
		}
	}
	h.loadHooks = append(h.loadHooks, hook)
	return true
}

// AddUnloadHook adds an UnloadHook function to the Hook.
func (h *Hook) AddUnloadHook(hook symbol.UnloadHook) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, h := range h.unloadHooks {
		if h == hook {
			return false
		}
	}
	h.unloadHooks = append(h.unloadHooks, hook)
	return true
}

// Load executes all LoadHooks registered in the Hook on the provided symbol.
func (h *Hook) Load(sb *symbol.Symbol) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.loadHooks.Load(sb)
}

// Unload executes all UnloadHooks registered in the Hook on the provided symbol.
func (h *Hook) Unload(sb *symbol.Symbol) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.unloadHooks.Unload(sb)
}
