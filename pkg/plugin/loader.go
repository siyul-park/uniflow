package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"io"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
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

	b, wrappers, err := l.wrap(i, b)
	if err != nil {
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

	methods := map[string]reflect.Value{}
	for name, wrapper := range wrappers {
		val, err := i.Eval("main." + wrapper)
		if err != nil {
			return nil, err
		}
		if val.Kind() != reflect.Func {
			return nil, errors.WithStack(ErrInvalidSignature)
		}
		methods[name] = val
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

func (l *Loader) wrap(i *interp.Interpreter, b []byte) ([]byte, map[string]string, error) {
	var buf bytes.Buffer
	buf.Write(b)

	node, err := parser.ParseFile(i.FileSet(), ".go", b, parser.AllErrors)
	if err != nil {
		return nil, nil, err
	}

	var ctor *ast.FuncDecl
	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "New" {
			continue
		}
		ctor = fn
		break
	}
	if ctor == nil {
		return nil, nil, errors.WithStack(ErrInvalidSignature)
	}

	var recv *ast.Ident
	if ctor.Type.Results != nil && len(ctor.Type.Results.List) > 0 {
		typExpr := ctor.Type.Results.List[0].Type
		switch t := typExpr.(type) {
		case *ast.StarExpr:
			if ident, ok := t.X.(*ast.Ident); ok {
				recv = ident
			}
		case *ast.Ident:
			recv = t
		}
	}
	if recv == nil {
		return nil, nil, errors.WithStack(ErrInvalidSignature)
	}

	methods := make(map[string]*ast.FuncDecl)
	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}

		var r *ast.Ident
		switch rt := fn.Recv.List[0].Type.(type) {
		case *ast.StarExpr:
			if id, ok := rt.X.(*ast.Ident); ok {
				r = id
			}
		case *ast.Ident:
			r = rt
		}

		if r != nil && r.Name == recv.Name {
			methods[fn.Name.Name] = fn
		}
	}

	wrappers := map[string]string{}
	for name, fn := range methods {

		params := []string{"recv *" + recv.Name}
		var args []string

		if fn.Type.Params != nil {
			for _, field := range fn.Type.Params.List {
				var typ bytes.Buffer
				if err := printer.Fprint(&typ, i.FileSet(), field.Type); err != nil {
					return nil, nil, err
				}

				count := len(field.Names)
				if count == 0 {
					count = 1
				}

				for i := 0; i < count; i++ {
					arg := fmt.Sprintf("a%d", len(args))
					params = append(params, fmt.Sprintf("%s %s", arg, typ.String()))
					args = append(args, arg)
				}
			}
		}

		ret := ""
		if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
			var rets []string
			for _, field := range fn.Type.Results.List {
				var typ bytes.Buffer
				err := printer.Fprint(&typ, i.FileSet(), field.Type)
				if err != nil {
					return nil, nil, err
				}
				count := 1
				if len(field.Names) > 0 {
					count = len(field.Names)
				}
				for i := 0; i < count; i++ {
					rets = append(rets, typ.String())
				}
			}
			if len(rets) == 1 {
				ret = rets[0]
			} else {
				ret = "(" + strings.Join(rets, ", ") + ")"
			}
		}

		call := fmt.Sprintf("recv.%s(%s)", name, strings.Join(args, ", "))

		var body string
		if ret == "" {
			body = call
		} else {
			body = "return " + call
		}

		wrapper := fmt.Sprintf(
			"func __wrap_%s_%s(%s) %s {\n\t%s\n}\n",
			recv.Name,
			name,
			strings.Join(params, ", "),
			ret,
			body,
		)

		buf.WriteString(wrapper)
		wrappers[name] = fmt.Sprintf("__wrap_%s_%s", recv.Name, name)
	}

	return buf.Bytes(), wrappers, nil
}
