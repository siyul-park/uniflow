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

func TestBreakpoint_Next(t *testing.T) {
	proc := process.New()
	sym := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}

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
