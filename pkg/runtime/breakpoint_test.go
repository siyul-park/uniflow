package runtime

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

	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	defer sb.Close()

	b := NewBreakpoint(
		BreakWithProcess(proc),
		BreakWithSymbol(sb),
		BreakWithInPort(sb.In(node.PortIn)),
		BreakWithOutPort(sb.Out(node.PortOut)),
	)
	defer b.Close()

	assert.NotZero(t, proc.ID())
	assert.Equal(t, proc, b.Process())
	assert.Equal(t, sb, b.Symbol())
	assert.Equal(t, sb.In(node.PortIn), b.InPort())
	assert.Equal(t, sb.Out(node.PortOut), b.OutPort())
}

func TestBreakpoint_Next(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	defer sb.Close()

	b := NewBreakpoint(
		BreakWithProcess(proc),
		BreakWithSymbol(sb),
	)
	defer b.Close()

	frame := &Frame{
		Process: proc,
		Symbol:  sb,
	}

	go b.OnFrame(frame)

	assert.True(t, b.Next())
	assert.Equal(t, frame, b.Frame())
}

func TestBreakpoint_Done(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	defer sb.Close()

	b := NewBreakpoint(
		BreakWithProcess(proc),
		BreakWithSymbol(sb),
	)
	defer b.Close()

	frame := &Frame{
		Process: proc,
		Symbol:  sb,
	}

	go b.OnFrame(frame)

	assert.True(t, b.Next())
	assert.Equal(t, frame, b.Frame())

	assert.True(t, b.Done())
	assert.Nil(t, b.Frame())
}
