package port

import (
	"github.com/siyul-park/uniflow/pkg/process"
)

// Handler is an interface that defines the Serve method to handle processes.
type Handler interface {
	Serve(proc *process.Process)
}

// HandlerFunc is an adapter to allow the use of ordinary functions as Handlers.
type HandlerFunc func(proc *process.Process)

var _ Handler = (HandlerFunc)(nil)

// Serve calls the underlying function for the HandlerFunc.
func (h HandlerFunc) Serve(proc *process.Process) {
	h(proc)
}
