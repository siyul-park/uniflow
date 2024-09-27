package agent

import "github.com/siyul-park/uniflow/pkg/process"

// Watcher defines methods for handling Frame and Process events.
type Watcher interface {
	OnFrame(*Frame)             // Called on Frame events.
	OnProcess(*process.Process) // Called on Process events.
}

type watcher struct {
	onFrame   func(*Frame)
	onProcess func(*process.Process)
}

var _ Watcher = (*watcher)(nil)

// NewFrameWatcher returns a Watcher for Frame events.
func NewFrameWatcher(handle func(*Frame)) Watcher {
	return &watcher{onFrame: handle}
}

// NewProcessWatcher returns a Watcher for Process events.
func NewProcessWatcher(handle func(*process.Process)) Watcher {
	return &watcher{onProcess: handle}
}

func (w *watcher) OnFrame(frame *Frame) {
	if w.onFrame != nil {
		w.onFrame(frame)
	}
}

func (w *watcher) OnProcess(process *process.Process) {
	if w.onProcess != nil {
		w.onProcess(process)
	}
}
