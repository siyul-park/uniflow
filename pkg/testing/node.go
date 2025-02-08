package testing

import (
	"time"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

type testNode struct {
	*node.OneToOneNode
	*BaseSuite
	spec     *TestNodeSpec
	scheme   *scheme.Scheme
	result   *Result
	inPorts  map[string]*port.InPort
	outPorts map[string]*port.OutPort
}

var _ node.Node = (*testNode)(nil)
var _ Suite = (*testNode)(nil)

// NewTestNode creates a new test node.
func NewTestNode(spec *TestNodeSpec, s *scheme.Scheme) (node.Node, error) {
	// Set default namespace if not specified
	if spec.GetNamespace() == "" {
		spec.SetNamespace(resource.DefaultNamespace)
	}

	// Ensure all child specs have namespace set
	for _, childSpec := range spec.Specs {
		if childSpec.GetNamespace() == "" {
			childSpec.SetNamespace(spec.GetNamespace())
		}
	}

	n := &testNode{
		spec:   spec,
		scheme: s,
		result: &Result{
			Name:      spec.Name,
			StartTime: time.Now(),
		},
		inPorts:  make(map[string]*port.InPort),
		outPorts: make(map[string]*port.OutPort),
	}

	// Create the OneToOneNode with process function
	n.OneToOneNode = node.NewOneToOneNode(n.process)

	// Create the BaseSuite with run function
	n.BaseSuite = NewSuite(n.runTest)

	// Configure ports based on spec
	for name, ports := range spec.Ports {
		for _, p := range ports {
			if p.Port == "in" {
				n.inPorts[name] = port.NewIn()
			} else if p.Port == "out" {
				n.outPorts[name] = port.NewOut()
			}
		}
	}

	return n, nil
}

func (n *testNode) runTest(t *Tester) {
	// Execute each spec in sequence
	for _, spec := range n.spec.Specs {
		// Create a new node for the spec
		node, err := n.scheme.Compile(spec)
		if err != nil {
			n.result.Error = err
			n.result.Status = StatusFailed
			n.result.EndTime = time.Now()
			t.Close(err)
			return
		}

		defer node.Close()

		// Connect the node's ports
		in := node.In("in")
		if in != nil {
			out := port.NewOut()
			defer out.Close()

			out.Link(in)
			writer := out.Open(t.Process)

			writer.Write(packet.New(nil)) // Create empty packet for initial test
			if _, ok := <-writer.Receive(); ok {
				// Continue processing
			}
		}
	}

	// Set final result status
	if n.result.Error == nil {
		n.result.Status = StatusPassed
	}
	n.result.EndTime = time.Now()
	t.Close(n.result.Error)
}

func (n *testNode) process(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	// The actual test execution happens in runTest
	// This is just a pass-through for the node interface
	return inPck, nil
}

func (n *testNode) GetResult() *Result {
	return n.result
}

// In returns the input port with the given name
func (n *testNode) In(name string) *port.InPort {
	if p, ok := n.inPorts[name]; ok {
		return p
	}
	return n.OneToOneNode.In(name)
}

// Out returns the output port with the given name
func (n *testNode) Out(name string) *port.OutPort {
	if p, ok := n.outPorts[name]; ok {
		return p
	}
	return n.OneToOneNode.Out(name)
}

// Close implements the Node interface
func (n *testNode) Close() error {
	// Close all custom ports
	for _, p := range n.inPorts {
		p.Close()
	}
	for _, p := range n.outPorts {
		p.Close()
	}
	return n.OneToOneNode.Close()
}
