package harness

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestSimpleFlow(t *testing.T) {
	proc := process.New()
	resultChan := make(chan *packet.Packet, 1)
	defer close(resultChan)

	// Create nodes with improved error handling
	source := NewSourceNode(types.NewString("hello world"))
	transform := NewTransformNode(func(v types.Value) (types.Value, error) {
		str, ok := v.(types.String)
		if !ok {
			return nil, fmt.Errorf("expected string value, got %T", v)
		}
		return types.NewString(strings.ToUpper(str.String())), nil
	})
	validator := NewAssertNode(resultChan, func(v types.Value) error {
		str, ok := v.(types.String)
		if !ok {
			return fmt.Errorf("expected string value, got %T", v)
		}
		if str.String() != "HELLO WORLD" {
			return fmt.Errorf("expected 'HELLO WORLD', got '%s'", str.String())
		}
		return nil
	})

	// Setup cleanup
	defer func() {
		source.Close()
		transform.Close()
		validator.Close()
	}()

	// Connect nodes
	source.Out(node.PortOut).Link(transform.In(node.PortIn))
	transform.Out(node.PortOut).Link(validator.In(node.PortIn))

	// Start the flow
	writer := packet.NewWriter()
	reader := source.In(node.PortIn).Open(proc)
	writer.Link(reader)
	defer writer.Close()

	// Trigger the flow
	packet.Send(writer, packet.New(types.NewString("trigger")))

	// Wait for and verify result with timeout
	select {
	case result := <-resultChan:
		assert.NotNil(t, result)
		payload := result.Payload()
		str, ok := payload.(types.String)
		assert.True(t, ok, "expected string payload")
		assert.Equal(t, "HELLO WORLD", str.String())
	case <-time.After(time.Second):
		t.Fatal("test timed out")
	}
}
