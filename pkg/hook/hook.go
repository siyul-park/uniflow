package hook

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/chart"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Hook represents a collection of hook functions that can be executed on charts and symbols.
type Hook struct {
	linkHooks   chart.LinkHooks
	unlinkHooks chart.UnlinkHooks
	loadHooks   symbol.LoadHooks
	unloadHooks symbol.UnloadHooks
	mu          sync.RWMutex
}

var _ chart.LinkHook = (*Hook)(nil)
var _ chart.UnlinkHook = (*Hook)(nil)
var _ symbol.LoadHook = (*Hook)(nil)
var _ symbol.UnloadHook = (*Hook)(nil)

// New creates a new instance of Hook.
func New() *Hook {
	return &Hook{}
}

// AddLinkHook adds a LinkHook function to the Hook.
func (h *Hook) AddLinkHook(hook chart.LinkHook) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.linkHooks = append(h.linkHooks, hook)
}

// AddUnlinkHook adds an UnlinkHook function to the Hook.
func (h *Hook) AddUnlinkHook(hook chart.UnlinkHook) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.unlinkHooks = append(h.unlinkHooks, hook)
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

// Link executes all LinkHooks registered in the Hook on the provided chart.
func (h *Hook) Link(chrt *chart.Chart) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.linkHooks.Link(chrt)
}

// Unlink executes all UnlinkHooks registered in the Hook on the provided chart.
func (h *Hook) Unlink(chrt *chart.Chart) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.unlinkHooks.Unlink(chrt)
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
