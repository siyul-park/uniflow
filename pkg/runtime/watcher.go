package runtime

import (
	"github.com/siyul-park/uniflow/pkg/process"
)

// Watcher defines methods for handling Frame and Process events.
type Watcher interface {
	// OnFrame is triggered when a Frame event occurs.
	OnFrame(*Frame)
	// OnProcess is triggered when a Process event occurs.
	OnProcess(*process.Process)
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

// OnFrame triggers the OnFrame method for each Watcher in the slice.
func (w Watchers) OnFrame(frame *Frame) {
	for _, watcher := range w {
		watcher.OnFrame(frame)
	}
}

// OnProcess triggers the OnProcess method for each Watcher in the slice.
func (w Watchers) OnProcess(proc *process.Process) {
	for _, watcher := range w {
		watcher.OnProcess(proc)
	}
}

// OnFrame triggers the onFrame function if it is defined.
func (w *watcher) OnFrame(frame *Frame) {
	if w.onFrame != nil {
		w.onFrame(frame)
	}
}

// OnProcess triggers the onProcess function if it is defined.
func (w *watcher) OnProcess(proc *process.Process) {
	if w.onProcess != nil {
		w.onProcess(proc)
	}
}
