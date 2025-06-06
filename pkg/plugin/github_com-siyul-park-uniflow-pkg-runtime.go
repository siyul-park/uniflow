// Code generated by 'yaegi extract github.com/siyul-park/uniflow/pkg/runtime'. DO NOT EDIT.

package plugin

import (
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"reflect"
)

func init() {
	Symbols["github.com/siyul-park/uniflow/pkg/runtime/runtime"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"BreakWithInPort":   reflect.ValueOf(runtime.BreakWithInPort),
		"BreakWithOutPort":  reflect.ValueOf(runtime.BreakWithOutPort),
		"BreakWithProcess":  reflect.ValueOf(runtime.BreakWithProcess),
		"BreakWithSymbol":   reflect.ValueOf(runtime.BreakWithSymbol),
		"New":               reflect.ValueOf(runtime.New),
		"NewAgent":          reflect.ValueOf(runtime.NewAgent),
		"NewBreakpoint":     reflect.ValueOf(runtime.NewBreakpoint),
		"NewDebugger":       reflect.ValueOf(runtime.NewDebugger),
		"NewFrameWatcher":   reflect.ValueOf(runtime.NewFrameWatcher),
		"NewProcessWatcher": reflect.ValueOf(runtime.NewProcessWatcher),

		// type definitions
		"Agent":      reflect.ValueOf((*runtime.Agent)(nil)),
		"Breakpoint": reflect.ValueOf((*runtime.Breakpoint)(nil)),
		"Config":     reflect.ValueOf((*runtime.Config)(nil)),
		"Debugger":   reflect.ValueOf((*runtime.Debugger)(nil)),
		"Frame":      reflect.ValueOf((*runtime.Frame)(nil)),
		"Runtime":    reflect.ValueOf((*runtime.Runtime)(nil)),
		"Watcher":    reflect.ValueOf((*runtime.Watcher)(nil)),
		"Watchers":   reflect.ValueOf((*runtime.Watchers)(nil)),

		// interface wrapper definitions
		"_Watcher": reflect.ValueOf((*_github_com_siyul_park_uniflow_pkg_runtime_Watcher)(nil)),
	}
}

// _github_com_siyul_park_uniflow_pkg_runtime_Watcher is an interface wrapper for Watcher type
type _github_com_siyul_park_uniflow_pkg_runtime_Watcher struct {
	IValue     interface{}
	WOnFrame   func(a0 *runtime.Frame)
	WOnProcess func(a0 *process.Process)
}

func (W _github_com_siyul_park_uniflow_pkg_runtime_Watcher) OnFrame(a0 *runtime.Frame) {
	W.WOnFrame(a0)
}
func (W _github_com_siyul_park_uniflow_pkg_runtime_Watcher) OnProcess(a0 *process.Process) {
	W.WOnProcess(a0)
}
