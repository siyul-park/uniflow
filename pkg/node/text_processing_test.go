package node

import (
	"strings"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestSimpleWorkflow(t *testing.T) {
	proc := process.New()
	resultChan := make(chan *packet.Packet, 1)

	// Create a simple source -> sink workflow
	source := NewOneToOneNode(func(p *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
		// Always generate output when receiving input
		if in == nil {
			return nil, nil
		}
		out := packet.New(types.NewString("test message"))
		return out, nil
	})

	sink := NewOneToOneNode(func(p *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
		if in != nil {
			resultChan <- in
		}
		return nil, nil
	})

	// Connect source to sink
	source.Out(PortOut).Link(sink.In(PortIn))

	// Open source input port
	sourceIn := source.In(PortIn)
	if sourceIn == nil {
		t.Fatal("Failed to get source input port")
	}

	// Create and link writer
	writer := packet.NewWriter()
	reader := sourceIn.Open(proc)
	writer.Link(reader)

	// Send initial packet
	initialPacket := packet.New(types.NewString("start"))
	if initialPacket == nil {
		t.Fatal("Failed to create initial packet")
	}
	packet.Send(writer, initialPacket)

	// Wait for result with timeout
	select {
	case result := <-resultChan:
		assert.NotNil(t, result)
		assert.Equal(t, "test message", result.Payload().(types.String).String())
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out waiting for result")
	}

	// Cleanup
	close(resultChan)
	writer.Close()
	source.Close()
	sink.Close()
}

func TestTextProcessingWorkflow(t *testing.T) {
	proc := process.New()
	resultChan := make(chan *packet.Packet, 1)

	// Create text processing nodes
	source := NewOneToOneNode(func(p *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
		if in == nil {
			return nil, nil
		}
		out := packet.New(types.NewString("Hello, World! This is a TEST message."))
		return out, nil
	})

	uppercaseNode := NewOneToOneNode(func(p *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
		if in == nil {
			return nil, nil
		}
		text := in.Payload().(types.String)
		out := packet.New(types.NewString(text.String()))
		return out, nil
	})

	sink := NewOneToOneNode(func(p *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
		if in != nil {
			resultChan <- in
		}
		return nil, nil
	})

	// Connect nodes
	source.Out(PortOut).Link(uppercaseNode.In(PortIn))
	uppercaseNode.Out(PortOut).Link(sink.In(PortIn))

	// Open source input port
	sourceIn := source.In(PortIn)
	if sourceIn == nil {
		t.Fatal("Failed to get source input port")
	}

	// Create and link writer
	writer := packet.NewWriter()
	reader := sourceIn.Open(proc)
	writer.Link(reader)

	// Send initial packet
	initialPacket := packet.New(types.NewString("start"))
	if initialPacket == nil {
		t.Fatal("Failed to create initial packet")
	}
	packet.Send(writer, initialPacket)

	// Wait for result with timeout
	select {
	case result := <-resultChan:
		assert.NotNil(t, result)
		text := result.Payload().(types.String).String()
		assert.Equal(t, "Hello, World! This is a TEST message.", text)
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out waiting for result")
	}

	// Cleanup
	close(resultChan)
	writer.Close()
	source.Close()
	uppercaseNode.Close()
	sink.Close()
}

// Example of a text filtering workflow
func TestTextFilteringWorkflow(t *testing.T) {
	proc := process.New()
	resultChan := make(chan *packet.Packet, 3)

	// Source node with multiple sentences
	source := NewOneToOneNode(func(p *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
		// Always generate data, regardless of input
		return packet.New(types.NewString("First line.\nSecond line with ERROR.\nThird line is OK.\nFourth line has WARNING.")), nil
	})

	// Split text into lines
	splitter := NewOneToManyNode(func(p *process.Process, in *packet.Packet) ([]*packet.Packet, *packet.Packet) {
		if in == nil {
			return nil, nil
		}
		text := in.Payload().(types.String)
		lines := strings.Split(text.String(), "\n")

		// Create a packet for each line
		packets := make([]*packet.Packet, len(lines))
		for i, line := range lines {
			packets[i] = packet.New(types.NewString(line))
		}
		return packets, nil
	})

	// Filter lines containing specific keywords
	errorFilter := NewOneToOneNode(func(p *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
		if in == nil {
			return nil, nil
		}
		text := in.Payload().(types.String)
		if strings.Contains(text.String(), "ERROR") {
			return packet.New(types.NewString("[ERROR] " + text.String())), nil
		}
		return nil, nil
	})

	warningFilter := NewOneToOneNode(func(p *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
		if in == nil {
			return nil, nil
		}
		text := in.Payload().(types.String)
		if strings.Contains(text.String(), "WARNING") {
			return packet.New(types.NewString("[WARNING] " + text.String())), nil
		}
		return nil, nil
	})

	// Collect filtered results
	sink := NewOneToOneNode(func(p *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
		if in != nil {
			resultChan <- in
		}
		return nil, nil
	})

	// Connect the workflow
	source.Out(PortOut).Link(splitter.In(PortIn))
	splitter.Out(PortOut).Link(errorFilter.In(PortIn))
	splitter.Out(PortOut).Link(warningFilter.In(PortIn))
	errorFilter.Out(PortOut).Link(sink.In(PortIn))
	warningFilter.Out(PortOut).Link(sink.In(PortIn))

	// Start the workflow
	writer := packet.NewWriter()
	reader := source.In(PortIn).Open(proc)
	writer.Link(reader)

	// Send initial packet
	packet.Send(writer, packet.New(types.NewString("start")))

	// Collect results with timeout
	results := make([]string, 0)
	timeout := time.After(2 * time.Second)

	// We expect 2 results (one ERROR and one WARNING)
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
	assert.Equal(t, 2, len(results))
	assert.Contains(t, results, "[ERROR] Second line with ERROR.")
	assert.Contains(t, results, "[WARNING] Fourth line has WARNING.")

	// Cleanup
	close(resultChan)
	writer.Close()
	source.Close()
	splitter.Close()
	errorFilter.Close()
	warningFilter.Close()
	sink.Close()
}
