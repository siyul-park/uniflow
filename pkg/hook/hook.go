package hook

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

type (
	// Hook is a collection of hook functions.
	Hook struct {
		loadHooks   []symbol.LoadHook
		unloadHooks []symbol.UnloadHook
		mu          sync.RWMutex
	}
)

var _ symbol.LoadHook = &Hook{}
var _ symbol.UnloadHook = &Hook{}

// New returns a new Hooks.
func New() *Hook {
	return &Hook{}
}

// AddLoadHook adds a LoadHook.
func (h *Hook) AddLoadHook(hook symbol.LoadHook) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.loadHooks = append(h.loadHooks, hook)
}

// AddUnloadHook adds a UnloadHook.
func (h *Hook) AddUnloadHook(hook symbol.UnloadHook) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.unloadHooks = append(h.unloadHooks, hook)
}

// Load runs LoadHooks.
func (h *Hook) Load(n node.Node) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.loadHooks {
		hook.Load(n)
	}
}

// Unload runs UnloadHooks.
func (h *Hook) Unload(n node.Node) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.unloadHooks {
		hook.Unload(n)
	}
}
