package debug

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/assert"
)

func TestNewBreakpoint(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	sym := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	defer sym.Close()

	b := NewBreakpoint(
		WithProcess(proc),
		WithSymbol(sym),
		WithInPort(sym.In(node.PortIn)),
		WithOutPort(sym.Out(node.PortOut)),
	)
	defer b.Close()

	assert.Equal(t, proc, b.Process())
	assert.Equal(t, sym, b.Symbol())
	assert.Equal(t, sym.In(node.PortIn), b.InPort())
	assert.Equal(t, sym.Out(node.PortOut), b.OutPort())
}

func TestBreakpoint_Next(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	sym := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	defer sym.Close()

	b := NewBreakpoint(
		WithProcess(proc),
		WithSymbol(sym),
	)
	defer b.Close()

	frame := &Frame{
		Process: proc,
		Symbol:  sym,
	}

	go b.HandleFrame(frame)

	assert.True(t, b.Next())
	assert.Equal(t, frame, b.Frame())
}

func TestBreakpoint_Done(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	sym := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	defer sym.Close()

	b := NewBreakpoint(
		WithProcess(proc),
		WithSymbol(sym),
	)
	defer b.Close()

	frame := &Frame{
		Process: proc,
		Symbol:  sym,
	}

	go b.HandleFrame(frame)

	assert.True(t, b.Next())
	assert.Equal(t, frame, b.Frame())
	
	assert.True(t, b.Done())
	assert.Nil(t, b.Frame())
}
