package port

import (
	"sync"

	"github.com/siyul-park/uniflow/process"
)

// Listener is an interface for handling process events.
type Listener interface {
	// Accept is called to handle a process.
	Accept(proc *process.Process)
}

// Listeners is a slice of Listener interfaces, allowing multiple listeners to handle processes concurrently.
type Listeners []Listener

type listener struct {
	accept func(proc *process.Process)
}

var _ Listener = (Listeners)(nil)
var _ Listener = (*listener)(nil)

// ListenFunc creates a new Listener from the provided function.
func ListenFunc(accept func(proc *process.Process)) Listener {
	return &listener{accept: accept}
}

func (l Listeners) Accept(proc *process.Process) {
	wg := sync.WaitGroup{}
	for _, listener := range l {
		listener := listener
		wg.Add(1)
		go func() {
			defer wg.Done()
			listener.Accept(proc)
		}()
	}
	wg.Wait()
}

func (l *listener) Accept(proc *process.Process) {
	l.accept(proc)
}
