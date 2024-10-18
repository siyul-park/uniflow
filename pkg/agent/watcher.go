package agent

import (
	"github.com/siyul-park/uniflow/pkg/process"
)

// Watcher defines methods for handling Frame and Process events.
type Watcher interface {
	OnFrame(*Frame)             // Triggered when a Frame event occurs.
	OnProcess(*process.Process) // Triggered when a Process event occurs.
}

// Watchers is a slice of Watcher interfaces.
type Watchers []Watcher

type watcher struct {
	onFrame   func(*Frame)
	onProcess func(*process.Process)
}

var _ Watcher = (Watchers)(nil)
var _ Watcher = (*watcher)(nil)

// NewFrameWatcher creates a Watcher for handling Frame events.
func NewFrameWatcher(handle func(*Frame)) Watcher {
	return &watcher{onFrame: handle}
}

// NewProcessWatcher creates a Watcher for handling Process events.
func NewProcessWatcher(handle func(*process.Process)) Watcher {
	return &watcher{onProcess: handle}
}

func (w Watchers) OnFrame(frame *Frame) {
	for _, watcher := range w {
		watcher.OnFrame(frame)
	}
}

func (w Watchers) OnProcess(proc *process.Process) {
	for _, watcher := range w {
		watcher.OnProcess(proc)
	}
}

func (w *watcher) OnFrame(frame *Frame) {
	if w.onFrame != nil {
		w.onFrame(frame)
	}
}

func (w *watcher) OnProcess(proc *process.Process) {
	if w.onProcess != nil {
		w.onProcess(proc)
	}
}
