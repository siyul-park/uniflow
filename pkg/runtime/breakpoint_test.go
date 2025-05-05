package runtime

import (
	"encoding/json"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/pkg/meta"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

func TestNewBreakpoint(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: meta.DefaultNamespace,
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

	require.NotZero(t, b.ID())
	require.Equal(t, proc, b.Process())
	require.Equal(t, sb, b.Symbol())
	require.Equal(t, sb.In(node.PortIn), b.InPort())
	require.Equal(t, sb.Out(node.PortOut), b.OutPort())
}

func TestBreakpoint_Next(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: meta.DefaultNamespace,
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

	require.True(t, b.Next())
	require.Equal(t, frame, b.Frame())
}

func TestBreakpoint_Done(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: meta.DefaultNamespace,
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

	require.True(t, b.Next())
	require.Equal(t, frame, b.Frame())

	require.True(t, b.Done())
	require.Nil(t, b.Frame())
}

func TestBreakpoint_MarshalJSON(t *testing.T) {
	proc := process.New()
	defer proc.Exit(nil)

	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: meta.DefaultNamespace,
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

	data, err := json.Marshal(b)
	require.NoError(t, err)
	require.NotZero(t, data)
}
