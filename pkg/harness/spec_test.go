package harness

import (
	"fmt"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func setupTestWorkflow(t *testing.T, sourceNode, transformNode, assertNode node.Node, resultChan chan *packet.Packet) (*process.Process, *packet.Writer, func()) {
	proc := process.New()

	// Connect nodes
	if transformNode != nil {
		sourceNode.Out(node.PortOut).Link(transformNode.In(node.PortIn))
		transformNode.Out(node.PortOut).Link(assertNode.In(node.PortIn))
	} else {
		sourceNode.Out(node.PortOut).Link(assertNode.In(node.PortIn))
	}

	// Start the flow
	writer := packet.NewWriter()
	reader := sourceNode.In(node.PortIn).Open(proc)
	writer.Link(reader)

	cleanup := func() {
		proc.Exit(nil)
		proc.Join()
		writer.Close()
		if transformNode != nil {
			transformNode.Close()
		}
		sourceNode.Close()
		assertNode.Close()
		close(resultChan)
	}

	return proc, writer, cleanup
}

func TestSchemeBasedWorkflow(t *testing.T) {
	// Create a new scheme and register our node types
	s := scheme.New()
	err := AddToScheme().AddToScheme(s)
	assert.NoError(t, err)

	resultChan := make(chan *packet.Packet, 1)

	// Create test specs
	sourceSpec := &SourceNodeSpec{
		Meta: spec.Meta{
			Kind: KindSourceNode,
		},
		Data: "hello world",
	}

	transformSpec := &TransformNodeSpec{
		Meta: spec.Meta{
			Kind: KindTransformNode,
		},
		Transform: "return input.toUpperCase()",
	}

	assertSpec := &AssertNodeSpec{
		Meta: spec.Meta{
			Kind: KindAssertNode,
		},
		ExpectedValue: "HELLO WORLD",
		ResultChan:    resultChan,
	}

	// Create nodes from specs
	sourceNode, err := s.Compile(sourceSpec)
	assert.NoError(t, err)
	assert.NotNil(t, sourceNode)

	transformNode, err := s.Compile(transformSpec)
	assert.NoError(t, err)
	assert.NotNil(t, transformNode)

	assertNode, err := s.Compile(assertSpec)
	assert.NoError(t, err)
	assert.NotNil(t, assertNode)

	// Set up the workflow
	proc, writer, cleanup := setupTestWorkflow(t, sourceNode, transformNode, assertNode, resultChan)
	defer cleanup()

	// Trigger the flow
	packet.Send(writer, packet.New(types.NewString("trigger")))

	// Wait for result with timeout
	select {
	case result := <-resultChan:
		assert.NotNil(t, result)
		payload := result.Payload()
		str, ok := payload.(types.String)
		assert.True(t, ok, "expected string payload")
		assert.Equal(t, "HELLO WORLD", str.String())
	case <-proc.Done():
		t.Fatal("process exited before completion")
	case <-time.After(time.Second):
		t.Fatal("test timed out")
	}
}

func TestHarnessNodesWithDifferentTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "string value",
			input:    "test string",
			expected: "test string",
		},
		{
			name:     "boolean value",
			input:    true,
			expected: true,
		},
		{
			name:     "number as string",
			input:    42,
			expected: "42",
		},
		{
			name:     "complex type as string",
			input:    struct{ name string }{"test"},
			expected: "{test}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultChan := make(chan *packet.Packet, 1)

			// Create source with test input
			sourceSpec := &SourceNodeSpec{
				Meta: spec.Meta{Kind: KindSourceNode},
				Data: tt.input,
			}

			// Create assert with expected output
			assertSpec := &AssertNodeSpec{
				Meta:          spec.Meta{Kind: KindAssertNode},
				ExpectedValue: tt.expected,
				ResultChan:    resultChan,
			}

			s := scheme.New()
			err := AddToScheme().AddToScheme(s)
			assert.NoError(t, err)

			sourceNode, err := s.Compile(sourceSpec)
			assert.NoError(t, err)
			assert.NotNil(t, sourceNode)

			assertNode, err := s.Compile(assertSpec)
			assert.NoError(t, err)
			assert.NotNil(t, assertNode)

			// Set up the workflow
			proc, writer, cleanup := setupTestWorkflow(t, sourceNode, nil, assertNode, resultChan)
			defer cleanup()

			// Trigger the flow
			packet.Send(writer, packet.New(types.NewString("trigger")))

			// Wait for result
			select {
			case result := <-resultChan:
				assert.NotNil(t, result)
				payload := result.Payload()
				switch v := tt.expected.(type) {
				case string:
					str, ok := payload.(types.String)
					assert.True(t, ok, "expected string payload")
					assert.Equal(t, v, str.String())
				case bool:
					b, ok := payload.(types.Boolean)
					assert.True(t, ok, "expected boolean payload")
					assert.Equal(t, v, b.Bool())
				default:
					str, ok := payload.(types.String)
					assert.True(t, ok, "expected string payload")
					assert.Equal(t, fmt.Sprintf("%v", v), str.String())
				}
			case <-proc.Done():
				t.Fatal("process exited before completion")
			case <-time.After(time.Second):
				t.Fatal("test timed out")
			}
		})
	}
}
