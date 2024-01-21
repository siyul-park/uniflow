package js

import (
	"reflect"

	"github.com/dop251/goja"
	"github.com/iancoleman/strcase"
)

type fieldNameMapper struct{}

var _ goja.FieldNameMapper = &fieldNameMapper{}

func New() *goja.Runtime {
	vm := goja.New()
	vm.SetFieldNameMapper(&fieldNameMapper{})
	_ = UseModule(vm)
	return vm
}

func (*fieldNameMapper) FieldName(_ reflect.Type, f reflect.StructField) string {
	return strcase.ToLowerCamel(f.Name)
}

func (*fieldNameMapper) MethodName(_ reflect.Type, m reflect.Method) string {
	return strcase.ToLowerCamel(m.Name)
}
