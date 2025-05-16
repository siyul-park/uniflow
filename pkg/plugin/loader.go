package plugin

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"
)

type Loader struct {
	fs       afero.Fs
	validate *validator.Validate
}

func NewLoader(fs afero.Fs) *Loader {
	return &Loader{
		fs:       fs,
		validate: validator.New(validator.WithRequiredStructEnabled()),
	}
}

var ErrInvalidSignature = errors.New("invalid signature")

func (l *Loader) Open(path string, config any) (Plugin, error) {
	switch ext := filepath.Ext(path); ext {
	case ".so":
		return l.openNative(path, config)
	case ".go":
		return l.openInterp(path, config)
	default:
		return nil, errors.WithStack(ErrInvalidSignature)
	}
}

func (l *Loader) openNative(path string, config any) (Plugin, error) {
	tmp, err := os.CreateTemp("", "*.so")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())

	src, err := l.fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	if _, err := io.Copy(tmp, src); err != nil {
		tmp.Close()
		return nil, err
	}
	if err := tmp.Close(); err != nil {
		return nil, err
	}

	p, err := plugin.Open(tmp.Name())
	if err != nil {
		return nil, err
	}

	ctor, err := p.Lookup("New")
	if err != nil {
		return nil, err
	}

	recv, err := l.invoke(reflect.ValueOf(ctor), config)
	if err != nil {
		return nil, err
	}

	r, ok := recv.Interface().(Plugin)
	if !ok {
		return nil, ErrInvalidSignature
	}
	return r, nil
}

func (l *Loader) openInterp(path string, cfg any) (Plugin, error) {
	f, err := l.fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	i := interp.New(interp.Options{
		SourcecodeFilesystem: afero.NewIOFS(l.fs),
	})
	if err := i.Use(stdlib.Symbols); err != nil {
		return nil, err
	}
	if err := i.Use(Symbols); err != nil {
		return nil, err
	}

	node, err := parser.ParseFile(i.FileSet(), path, b, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	prg, err := i.CompileAST(node)
	if err != nil {
		return nil, err
	}
	if _, err := i.Execute(prg); err != nil {
		return nil, err
	}

	ctor, err := i.Eval("main.New")
	if err != nil {
		return nil, err
	}

	recv, err := l.invoke(ctor, cfg)
	if err != nil {
		return nil, err
	}

	if r, ok := recv.Interface().(Plugin); ok {
		return r, nil
	}

	conf := types.Config{Importer: importer.Default()}
	info := &types.Info{Defs: make(map[*ast.Ident]types.Object)}
	if _, err = conf.Check("main", i.FileSet(), []*ast.File{node}, info); err != nil {
		return nil, err
	}

	var fn *types.Func
	for id, obj := range info.Defs {
		if id.Name == "New" && obj != nil {
			if f, ok := obj.(*types.Func); ok {
				fn = f
				break
			}
		}
	}
	if fn == nil {
		return nil, errors.WithStack(ErrInvalidSignature)
	}

	sig := fn.Type().(*types.Signature)
	res := sig.Results()
	if res.Len() == 0 {
		return nil, errors.WithStack(ErrInvalidSignature)
	}

	typ := res.At(0).Type()
	if ptr, ok := typ.(*types.Pointer); ok {
		typ = ptr.Elem()
	}

	named, ok := typ.(*types.Named)
	if !ok {
		return nil, errors.WithStack(ErrInvalidSignature)
	}

	name := named.Obj().Name()
	methods := map[string]reflect.Value{}

	for j := 0; j < named.NumMethods(); j++ {
		m := named.Method(j)
		msig := m.Type().(*types.Signature)

		mname := m.Name()
		wname := fmt.Sprintf("__wrap_%s_%s", name, mname)

		params := []string{"recv *" + name}
		args := []string{"recv"}

		for h := 0; h < msig.Params().Len(); h++ {
			pn := fmt.Sprintf("a%d", h)
			pt := msig.Params().At(h).Type().String()

			params = append(params, fmt.Sprintf("%s %s", pn, pt))
			args = append(args, pn)
		}

		call := fmt.Sprintf("recv.%s(%s)", mname, strings.Join(args[1:], ", "))
		if msig.Results().Len() == 0 {
			call = call + "; return nil"
		} else {
			call = "return " + call
		}

		var rets []string
		for k := 0; k < msig.Results().Len(); k++ {
			rets = append(rets, msig.Results().At(k).Type().String())
		}

		var rblock string
		switch len(rets) {
		case 0:
			rblock = ""
		case 1:
			rblock = rets[0]
		default:
			rblock = "(" + strings.Join(rets, ", ") + ")"
		}

		var deps []types.Type
		for k := 0; k < msig.Params().Len(); k++ {
			deps = append(deps, msig.Params().At(k).Type())
		}
		for k := 0; k < msig.Results().Len(); k++ {
			deps = append(deps, msig.Results().At(k).Type())
		}

		imports := map[string]struct{}{}
		for len(deps) > 0 {
			curr := deps[0]
			deps = deps[1:]

			switch tt := curr.(type) {
			case *types.Named:
				obj := tt.Obj()
				if pkg := obj.Pkg(); pkg != nil && pkg.Path() != "main" {
					imports[pkg.Path()] = struct{}{}
				}
			case *types.Pointer:
				deps = append(deps, tt.Elem())
			case *types.Array:
				deps = append(deps, tt.Elem())
			case *types.Slice:
				deps = append(deps, tt.Elem())
			case *types.Map:
				deps = append(deps, tt.Key(), tt.Elem())
			case *types.Chan:
				deps = append(deps, tt.Elem())
			case *types.Signature:
				for i := 0; i < tt.Params().Len(); i++ {
					deps = append(deps, tt.Params().At(i).Type())
				}
				for i := 0; i < tt.Results().Len(); i++ {
					deps = append(deps, tt.Results().At(i).Type())
				}
			case *types.Struct:
				for i := 0; i < tt.NumFields(); i++ {
					deps = append(deps, tt.Field(i).Type())
				}
			}
		}

		var iblock string
		if len(imports) > 0 {
			var b strings.Builder
			b.WriteString("import (\n")
			for p := range imports {
				b.WriteString(fmt.Sprintf("\t%q\n", p))
			}
			b.WriteString(")\n")
			iblock = b.String()
		}

		if _, err := i.Eval(fmt.Sprintf(`
package main
%s
func %s(%s) %s {
	%s
}
`, iblock, wname, strings.Join(params, ", "), rblock, call)); err != nil {
			return nil, err
		}

		wrapper, err := i.Eval("main." + wname)
		if err != nil {
			return nil, err
		}
		if wrapper.Kind() != reflect.Func {
			return nil, errors.WithStack(ErrInvalidSignature)
		}

		methods[mname] = wrapper
	}

	return &proxy{
		receiver: recv,
		methods:  methods,
	}, nil
}

func (l *Loader) invoke(val reflect.Value, config any) (reflect.Value, error) {
	if val.Kind() != reflect.Func {
		return reflect.Value{}, errors.WithStack(ErrInvalidSignature)
	}

	var ins []reflect.Value
	for i := 0; i < val.Type().NumIn(); i++ {
		data, err := json.Marshal(config)
		if err != nil {
			return reflect.Value{}, err
		}
		in := reflect.New(val.Type().In(i))
		if err := json.Unmarshal(data, in.Interface()); err != nil {
			return reflect.Value{}, err
		}
		if err := l.validate.Struct(in.Interface()); err != nil {
			return reflect.Value{}, err
		}
		ins = append(ins, in.Elem())
	}

	res := val.Call(ins)
	if len(res) == 0 {
		return reflect.Value{}, errors.WithStack(ErrInvalidSignature)
	}

	if len(res) > 1 {
		if err, ok := res[len(res)-1].Interface().(error); ok && err != nil {
			return reflect.Value{}, err
		}
	}
	return res[0], nil
}
