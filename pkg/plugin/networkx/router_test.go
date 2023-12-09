package networkx

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewRouterNode(t *testing.T) {
	n := NewRouterNode()
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestRouterNode_Send(t *testing.T) {
	n := NewRouterNode()
	defer func() { _ = n.Close() }()

	in := port.New()
	inPort, _ := n.Port(node.PortIn)
	inPort.Link(in)

	n.Add(http.MethodGet, "/*", port.SetIndex(node.PortOut, 0))
	n.Add(http.MethodGet, "/:1/second", port.SetIndex(node.PortOut, 1))
	n.Add(http.MethodGet, "/:1/:2", port.SetIndex(node.PortOut, 2))

	var testCases = []struct {
		name         string
		whenURL      string
		expectPort   string
		expectParams map[string]string
	}{
		{
			name:         "route /first to /*",
			whenURL:      "/first",
			expectPort:   port.SetIndex(node.PortOut, 0),
			expectParams: map[string]string{"*": "first"},
		},
		{
			name:         "route /first/second to /:1/second",
			whenURL:      "/first/second",
			expectPort:   port.SetIndex(node.PortOut, 1),
			expectParams: map[string]string{"1": "first"},
		},
		{
			name:       "route /first/second-new to /:1/:2",
			whenURL:    "/first/second-new",
			expectPort: port.SetIndex(node.PortOut, 2),
			expectParams: map[string]string{
				"1": "first",
				"2": "second-new",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out := port.New()
			defer out.Close()
			outPort, _ := n.Port(tc.expectPort)
			outPort.Link(out)

			proc := process.New()
			defer proc.Exit(nil)

			inStream := in.Open(proc)
			outStream := out.Open(proc)

			inStream.Send(packet.New(primitive.NewMap(
				primitive.NewString(KeyMethod), primitive.NewString(http.MethodGet),
				primitive.NewString(KeyPath), primitive.NewString(tc.whenURL),
			)))

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			select {
			case outPck := <-outStream.Receive():
				outPayload, ok := outPck.Payload().(*primitive.Map)
				assert.True(t, ok)

				param, ok := outPayload.Get(primitive.NewString(KeyParams))
				assert.True(t, ok)
				method, ok := outPayload.Get(primitive.NewString(KeyMethod))
				assert.True(t, ok)
				path, ok := outPayload.Get(primitive.NewString(KeyPath))
				assert.True(t, ok)

				assert.Equal(t, tc.expectParams, param.Interface())
				assert.Equal(t, http.MethodGet, method.Interface())
				assert.Equal(t, tc.whenURL, path.Interface())
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}
		})
	}
}

func BenchmarkRouterNode_Send(b *testing.B) {
	n := NewRouterNode()
	defer func() { _ = n.Close() }()

	in := port.New()
	inPort, _ := n.Port(node.PortIn)
	inPort.Link(in)

	n.Add(http.MethodGet, "/*", port.SetIndex(node.PortOut, 0))
	n.Add(http.MethodGet, "/:1/second", port.SetIndex(node.PortOut, 1))
	n.Add(http.MethodGet, "/:1/:2", port.SetIndex(node.PortOut, 2))

	out := port.New()
	defer out.Close()
	outPort, _ := n.Port(port.SetIndex(node.PortOut, 0))
	outPort.Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	inStream := in.Open(proc)
	outStream := out.Open(proc)

	inPayload := primitive.NewMap(
		primitive.NewString(KeyMethod), primitive.NewString(http.MethodGet),
		primitive.NewString(KeyPath), primitive.NewString("/first"),
	)
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inStream.Send(inPck)
		<-outStream.Receive()
	}
}
