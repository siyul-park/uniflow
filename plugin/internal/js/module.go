package js

import (
	"github.com/dop251/goja"
)

const (
	keyModule  = "module"
	keyExports = "exports"
)

func UseModule(vm *goja.Runtime) {
	module := vm.NewObject()
	exports := vm.NewObject()

	_ = module.Set(keyExports, exports)

	_ = vm.Set(keyModule, module)
	_ = vm.Set(keyExports, exports)
}

func Export(vm *goja.Runtime, name string) goja.Value {
	module := vm.Get(keyModule)
	if module == nil {
		return nil
	}
	exports := module.ToObject(vm).Get(keyExports)
	if exports == nil {
		return nil
	}

	if name == "default" {
		exModule := exports.ToObject(vm).Get("__esModule")
		if exModule != nil && exModule.Export() == true {
			return exports.ToObject(vm).Get("default")
		} else {
			return exports
		}
	}
	return exports.ToObject(vm).Get(name)
}
