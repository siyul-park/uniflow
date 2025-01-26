package node

import (
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestIntegratedWorkflow(t *testing.T) {
	proc := process.New()

	// Create a channel to store the final result
	resultChan := make(chan *packet.Packet, 1)

	// Create nodes with appropriate actions
	source := NewOneToOneNode(func(p *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
		// Source generates data when it receives any input
		if in != nil {
			return packet.New(types.NewString("test_data")), nil
		}
		return nil, nil
	})

	transformer := NewOneToOneNode(func(p *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
		if in == nil {
			return nil, nil
		}
		// Transform the data
		data := in.Payload().(types.String)
		transformed := types.NewString(data.String() + "_transformed")
		return packet.New(transformed), nil
	})

	merger := NewOneToOneNode(func(p *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
		// Just pass through for now
		return in, nil
	})

	sink := NewOneToOneNode(func(p *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
		// Store the received packet in our channel
		if in != nil {
			resultChan <- in
		}
		return nil, nil
	})

	// Connect nodes in sequence
	source.Out(PortOut).Link(sink.In(PortIn))
	transformer.Out(PortOut).Link(merger.In(PortIn))
	merger.Out(PortOut).Link(sink.In(PortIn))

	// Start the process
	sourceIn := source.In(PortIn)
	writer := packet.NewWriter()
	writer.Link(sourceIn.Open(proc))

	// Trigger the workflow by sending an initial packet
	packet.Send(writer, packet.New(types.NewString("trigger")))

	// Wait for result with timeout
	var result *packet.Packet
	select {
	case result = <-resultChan:
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out waiting for result")
	}

	// Verify the result
	assert.NotNil(t, result)
	assert.Equal(t, "test_data_transformed", result.Payload().(types.String).String())

	// Cleanup
	close(resultChan)
	writer.Close()
	source.Close()
	transformer.Close()
	merger.Close()
	sink.Close()
}

// Example of a more complex workflow
func TestBranchingWorkflow(t *testing.T) {
	proc := process.New()
	resultChan := make(chan *packet.Packet, 2)

	// Source node that generates numbers
	source := NewOneToOneNode(func(p *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
		if in != nil {
			return packet.New(types.NewString("10")), nil
		}
		return nil, nil
	})

	// Two transformers for different operations
	doubler := NewOneToOneNode(func(p *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
		if in == nil {
			return nil, nil
		}
		data := in.Payload().(types.String)
		return packet.New(types.NewString("doubled_" + data.String())), nil
	})

	tripler := NewOneToOneNode(func(p *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
		if in == nil {
			return nil, nil
		}
		data := in.Payload().(types.String)
		return packet.New(types.NewString("tripled_" + data.String())), nil
	})

	// Merger that combines results
	merger := NewOneToOneNode(func(p *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
		if in == nil {
			return nil, nil
		}
		return in, nil
	})

	// Sink that collects results
	sink := NewOneToOneNode(func(p *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
		if in != nil {
			resultChan <- in
		}
		return nil, nil
	})

	// Connect the branching workflow
	source.Out(PortOut).Link(doubler.In(PortIn))
	source.Out(PortOut).Link(tripler.In(PortIn))
	doubler.Out(PortOut).Link(merger.In(PortIn))
	tripler.Out(PortOut).Link(merger.In(PortIn))
	merger.Out(PortOut).Link(sink.In(PortIn))

	// Start the workflow
	sourceIn := source.In(PortIn)
	writer := packet.NewWriter()
	writer.Link(sourceIn.Open(proc))

	// Send initial packet
	packet.Send(writer, packet.New(types.NewString("start")))

	// Collect results with timeout
	results := make([]string, 0, 2)
	timeout := time.After(2 * time.Second)

	for i := 0; i < 2; i++ {
		select {
		case pkt := <-resultChan:
			if pkt != nil {
				results = append(results, pkt.Payload().(types.String).String())
			}
		case <-timeout:
			t.Fatal("Test timed out waiting for results")
		}
	}

	// Verify results
	assert.Contains(t, results, "doubled_10")
	assert.Contains(t, results, "tripled_10")

	// Cleanup
	close(resultChan)
	writer.Close()
	source.Close()
	doubler.Close()
	tripler.Close()
	merger.Close()
	sink.Close()
}
