package plugin

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"io"
	"os"
	"path/filepath"
	"plugin"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// Loader loads plugins from Go source or compiled shared object files.
type Loader struct {
	fs       afero.Fs
	validate *validator.Validate
}

// LoadOptions specifies options for loading a plugin.
type LoadOptions struct {
	GoPath string
	Config any
}

// NewLoader returns a new Loader using the given filesystem.
func NewLoader(fs afero.Fs) *Loader {
	return &Loader{
		fs:       fs,
		validate: validator.New(validator.WithRequiredStructEnabled()),
	}
}

var ErrInvalidSignature = errors.New("invalid signature")

// Open loads and initializes a plugin with the given config.
func (l *Loader) Open(path string, options ...LoadOptions) (Plugin, error) {
	switch ext := filepath.Ext(path); ext {
	case ".so":
		return l.openNative(path, options...)
	case ".go":
		return l.openInterp(path, options...)
	default:
		return nil, errors.WithStack(ErrInvalidSignature)
	}
}

func (l *Loader) openNative(path string, options ...LoadOptions) (Plugin, error) {
	var config any
	for _, opt := range options {
		if opt.Config != nil {
			config = opt.Config
		}
	}

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

func (l *Loader) openInterp(path string, options ...LoadOptions) (Plugin, error) {
	var gopath string
	var config any
	for _, opt := range options {
		if opt.GoPath != "" {
			gopath = opt.GoPath
		}
		if opt.Config != nil {
			config = opt.Config
		}
	}

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
		GoPath:               gopath,
		Env:                  os.Environ(),
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

	fn := l.funcDecl(node, "New")
	if fn == nil {
		return nil, errors.WithStack(ErrInvalidSignature)
	}

	wrappers := map[string]string{}
	if r0 := l.result(fn, 0); r0 != nil {
		for _, method := range l.methods(node, r0) {
			if method.Name == nil {
				continue
			}
			wrapper := l.wrapper(method)
			if wrapper == nil {
				continue
			}
			node.Decls = append(node.Decls, wrapper)
			wrappers[method.Name.Name] = wrapper.Name.Name
		}
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

	recv, err := l.invoke(ctor, config)
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

func (l *Loader) invoke(fn reflect.Value, args ...any) (reflect.Value, error) {
	if fn.Kind() != reflect.Func {
		return reflect.Value{}, errors.WithStack(ErrInvalidSignature)
	}

	var ins []reflect.Value
	for i := 0; i < fn.Type().NumIn(); i++ {
		in := reflect.New(fn.Type().In(i))

		if i < len(args) {
			data, err := json.Marshal(args[i])
			if err != nil {
				return reflect.Value{}, err
			}
			if err := json.Unmarshal(data, in.Interface()); err != nil {
				return reflect.Value{}, err
			}
		}

		if err := l.validate.Struct(in.Interface()); err != nil {
			return reflect.Value{}, err
		}
		ins = append(ins, in.Elem())
	}

	res := fn.Call(ins)
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

func (*Loader) funcDecl(node *ast.File, name string) *ast.FuncDecl {
	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Name.Name == name {
			return fn
		}
	}
	return nil
}

func (*Loader) result(fn *ast.FuncDecl, index int) *ast.Ident {
	if fn.Type.Results == nil || index >= len(fn.Type.Results.List) {
		return nil
	}
	switch t := fn.Type.Results.List[index].Type.(type) {
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id
		}
	case *ast.Ident:
		return t
	}
	return nil
}

func (l *Loader) methods(file *ast.File, recv *ast.Ident) []*ast.FuncDecl {
	var methods []*ast.FuncDecl
	for _, decl := range file.Decls {
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
			methods = append(methods, fn)
		}
	}
	return methods
}

func (l *Loader) wrapper(method *ast.FuncDecl) *ast.FuncDecl {
	var recv *ast.Ident
	if method.Recv != nil && len(method.Recv.List) == 1 {
		switch rt := method.Recv.List[0].Type.(type) {
		case *ast.StarExpr:
			if id, ok := rt.X.(*ast.Ident); ok {
				recv = id
			}
		case *ast.Ident:
			recv = rt
		}
	}
	if recv == nil {
		return nil
	}

	var args []ast.Expr
	params := []*ast.Field{{Names: []*ast.Ident{{Name: "recv"}}, Type: &ast.StarExpr{X: recv}}}
	if method.Type.Params != nil {
		for _, field := range method.Type.Params.List {
			count := len(field.Names)
			if count == 0 {
				count = 1
			}
			for i := 0; i < count; i++ {
				arg := ast.NewIdent(fmt.Sprintf("a%d", len(args)))
				param := &ast.Field{
					Names: []*ast.Ident{arg},
					Type:  field.Type,
				}
				args = append(args, arg)
				params = append(params, param)
			}
		}
	}

	expr := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent("recv"),
			Sel: ast.NewIdent(method.Name.Name),
		},
		Args: args,
	}

	var stmt ast.Stmt
	if method.Type.Results != nil && len(method.Type.Results.List) > 0 {
		stmt = &ast.ReturnStmt{Results: []ast.Expr{expr}}
	} else {
		stmt = &ast.ExprStmt{X: expr}
	}

	return &ast.FuncDecl{
		Name: ast.NewIdent("__wrap_" + recv.Name + "_" + method.Name.Name),
		Type: &ast.FuncType{
			Params:  &ast.FieldList{List: params},
			Results: method.Type.Results,
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{stmt},
		},
	}
}
