package jshelper

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
)

func TestUseModuleAndExport(t *testing.T) {
	vm := goja.New()

	err := UseModule(vm)
	assert.NoError(t, err)

	_, err = vm.RunString("module.exports = {};")
	assert.NoError(t, err)

	v := GetExport(vm, "default")
	assert.NotNil(t, v)
}
