package js

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
)

func TestUseModuleAndExport(t *testing.T) {
	vm := goja.New()

	UseModule(vm)

	_, err := vm.RunString("module.exports = {};")
	assert.NoError(t, err)

	v := Export(vm, "default")
	assert.NotNil(t, v)
}

func TestAssertExportFunction(t *testing.T) {
	ok := AssertExportFunction("$", "default")
	assert.False(t, ok)

	ok = AssertExportFunction("module.exports = () => {};", "default")
	assert.True(t, ok)
}
