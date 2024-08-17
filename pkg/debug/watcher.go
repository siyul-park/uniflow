package debug

import "github.com/siyul-park/uniflow/pkg/process"

// Watcher interface with methods for handling *Frame and *Process.
type Watcher interface {
	HandleFrame(*Frame)
	HandleProcess(*process.Process)
}

type watcher struct {
	handleFrame   func(*Frame)
	handleProcess func(*process.Process)
}

var _ Watcher = (*watcher)(nil)

// HandleFrameFunc returns a Watcher that handles *Frame.
func HandleFrameFunc(handle func(*Frame)) Watcher {
	return &watcher{handleFrame: handle}
}

// HandleProcessFunc returns a Watcher that handles *Process.
func HandleProcessFunc(handle func(*process.Process)) Watcher {
	return &watcher{handleProcess: handle}
}

func (w *watcher) HandleFrame(frame *Frame) {
	if w.handleFrame != nil {
		w.handleFrame(frame)
	}
}

func (w *watcher) HandleProcess(process *process.Process) {
	if w.handleProcess != nil {
		w.handleProcess(process)
	}
}
