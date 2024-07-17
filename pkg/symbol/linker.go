package symbol

// import (
// 	"reflect"

// 	"github.com/gofrs/uuid"
// 	"github.com/siyul-park/uniflow/pkg/node"
// 	"github.com/siyul-park/uniflow/pkg/port"
// 	"github.com/siyul-park/uniflow/pkg/scheme"
// 	"github.com/siyul-park/uniflow/pkg/types"
// )

// type Linker struct {
// 	scheme *scheme.Scheme
// }

// var _ LoadHook = (*Linker)(nil)

// func NewLinker(scheme *scheme.Scheme) *Linker {
// 	return &Linker{scheme: scheme}
// }

// func (l *Linker) Load(sym *Symbol) error {
// 	value, err := l.init(sym)
// 	if err != nil {
// 		return err
// 	}

// 	if sym.Node != nil && reflect.DeepEqual(sym.Value, value) {
// 		return nil
// 	}

// 	if err := sym.Close(); err != nil {
// 		return err
// 	}

// 	sym.Value = value

// 	s, err := l.scheme.Decode(sym.Spec, value)
// 	if err != nil {
// 		return err
// 	}

// 	sym.Node, err = l.scheme.Compile(s)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (l *Linker) init(sym *Symbol) (any, error) {
// 	out := port.NewOut()
// 	defer out.Close()

// 	links := sym.Links()
// 	for _, location := range links[node.PortInit] {
// 		id := location.ID
// 		if id == (uuid.UUID{}) {
// 			id = t.lookup(sym.Namespace(), location.Name)
// 		}

// 		if ref, ok := t.symbols[id]; ok {
// 			if ref.Namespace() == sym.Namespace() {
// 				if in := ref.In(location.Port); in != nil {
// 					out.Link(in)
// 				}
// 			}
// 		}
// 	}

// 	payload, err := types.TextEncoder.Encode(sym.Spec)
// 	if err != nil {
// 		return nil, err
// 	}
// 	payload, err = port.Write(out, payload)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return types.InterfaceOf(payload), nil
// }
