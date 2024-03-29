package js

import (
	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
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

func AssertExportFunction(code, name string) bool {
	if v, err := Transform(code, api.TransformOptions{Loader: api.LoaderTS}); err == nil {
		code = v
	}
	if v, err := Transform(code, api.TransformOptions{Format: api.FormatCommonJS}); err == nil {
		code = v
	}

	program, err := goja.Compile("", code, true)
	if err != nil {
		return false
	}

	vm := New()
	if _, err := vm.RunProgram(program); err != nil {
		return false
	}

	if defaults := Export(vm, name); defaults == nil {
		return false
	} else if _, ok := goja.AssertFunction(defaults); !ok {
		return false
	}
	return true
}
