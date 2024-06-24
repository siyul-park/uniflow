package javascript

import "github.com/dop251/goja"

const (
	keyModule  = "module"
	keyExports = "exports"
	keyDefault = "default"
)

func useModule(vm *goja.Runtime) {
	module := vm.NewObject()
	exports := vm.NewObject()

	_ = module.Set(keyExports, exports)

	_ = vm.Set(keyModule, module)
	_ = vm.Set(keyExports, exports)
}

func export(vm *goja.Runtime, name string) goja.Value {
	module := vm.Get(keyModule)
	if module == nil {
		return nil
	}
	exports := module.ToObject(vm).Get(keyExports)
	if exports == nil {
		return nil
	}

	if name == keyDefault {
		exModule := exports.ToObject(vm).Get("__esModule")
		if exModule != nil && exModule.Export() == true {
			return exports.ToObject(vm).Get(keyDefault)
		} else {
			return exports
		}
	}
	return exports.ToObject(vm).Get(name)
}
