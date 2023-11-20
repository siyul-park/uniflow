package hook

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

type (
	// Hook is a collection of hook functions.
	Hook struct {
		preLoadHooks    []symbol.PreLoadHook
		postLoadHooks   []symbol.PostLoadHook
		preUnloadHooks  []symbol.PreUnloadHook
		postUnloadHooks []symbol.PostUnloadHook
		mu              sync.RWMutex
	}
)

var _ symbol.PreLoadHook = &Hook{}
var _ symbol.PostLoadHook = &Hook{}
var _ symbol.PreUnloadHook = &Hook{}
var _ symbol.PostUnloadHook = &Hook{}

// New returns a new Hooks.
func New() *Hook {
	return &Hook{}
}

// AddPreLoadHook adds a PreLoadHook.
func (h *Hook) AddPreLoadHook(hook symbol.PreLoadHook) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.preLoadHooks = append(h.preLoadHooks, hook)
}

// AddPostLoadHook adds a PostLoadHook.
func (h *Hook) AddPostLoadHook(hook symbol.PostLoadHook) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.postLoadHooks = append(h.postLoadHooks, hook)
}

// AddPreUnloadHook adds a PreUnloadHook.
func (h *Hook) AddPreUnloadHook(hook symbol.PreUnloadHook) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.preUnloadHooks = append(h.preUnloadHooks, hook)
}

// AddPostUnloadHook adds a PostUnloadHook.
func (h *Hook) AddPostUnloadHook(hook symbol.PostUnloadHook) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.postUnloadHooks = append(h.postUnloadHooks, hook)
}

// PreLoad runs PreLoadHooks.
func (h *Hook) PreLoad(n node.Node) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.preLoadHooks {
		if err := hook.PreLoad(n); err != nil {
			return err
		}
	}
	return nil
}

// PostLoad runs PostLoadHooks.
func (h *Hook) PostLoad(n node.Node) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.postLoadHooks {
		if err := hook.PostLoad(n); err != nil {
			return err
		}
	}
	return nil
}

// PreUnload runs PreUnloadHooks.
func (h *Hook) PreUnload(n node.Node) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.preUnloadHooks {
		if err := hook.PreUnload(n); err != nil {
			return err
		}
	}
	return nil
}

// PostUnload runs PostUnloadHooks.
func (h *Hook) PostUnload(n node.Node) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.postUnloadHooks {
		if err := hook.PostUnload(n); err != nil {
			return err
		}
	}
	return nil
}
