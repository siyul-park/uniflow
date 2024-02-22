package port

import (
	"github.com/siyul-park/uniflow/pkg/process"
)

type Handler interface {
	Serve(proc *process.Process)
}

type HandlerFunc func(proc *process.Process)

var _ Handler = (HandlerFunc)(nil)

func (h HandlerFunc) Serve(proc *process.Process) {
	h(proc)
}
