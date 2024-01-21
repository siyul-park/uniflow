package jsutil

import (
	"github.com/dop251/goja"
)

func UseModule(vm *goja.Runtime) error {
	module := vm.NewObject()
	exports := vm.NewObject()

	if err := vm.Set("module", module); err != nil {
		return err
	}
	if err := vm.Set("exports", exports); err != nil {
		return err
	}
	if err := module.Set("exports", exports); err != nil {
		return err
	}
	return nil
}

func GetExport(vm *goja.Runtime, name string) goja.Value {
	module := vm.Get("module")
	if module == nil {
		return nil
	}
	exports := module.ToObject(vm).Get("exports")
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
