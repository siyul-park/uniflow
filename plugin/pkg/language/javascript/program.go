package javascript

import (
	"reflect"
	"sync"

	"github.com/dop251/goja"
	"github.com/iancoleman/strcase"
	"github.com/siyul-park/uniflow/plugin/pkg/language"
)

type program struct {
	vms sync.Pool
}

type fieldNameMapper struct{}

var _ language.Program = (*program)(nil)
var _ goja.FieldNameMapper = &fieldNameMapper{}

func newProgram(prgrm *goja.Program) (language.Program, error) {
	p := &program{}
	p.vms.New = func() any {
		vm := goja.New()
		vm.SetFieldNameMapper(&fieldNameMapper{})
		useModule(vm)
		vm.RunProgram(prgrm)
		return vm
	}
	return p, nil
}

func (p *program) Run(env any) (any, error) {
	vm := p.vms.Get().(*goja.Runtime)
	defer p.vms.Put(vm)

	export := export(vm, keyDefault)
	run, ok := goja.AssertFunction(export)
	if !ok {
		return nil, nil
	}

	if result, err := run(goja.Undefined(), vm.ToValue(env)); err != nil {
		return nil, err
	} else {
		return result.Export(), nil
	}
}

func (*fieldNameMapper) FieldName(_ reflect.Type, f reflect.StructField) string {
	return strcase.ToLowerCamel(f.Name)
}

func (*fieldNameMapper) MethodName(_ reflect.Type, m reflect.Method) string {
	return strcase.ToLowerCamel(m.Name)
}
