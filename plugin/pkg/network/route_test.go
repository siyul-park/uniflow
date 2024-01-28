package network

import (
	"context"
	"fmt"
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

func TestNewRouteNode(t *testing.T) {
	n := NewRouteNode()
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestRouteNode_Add(t *testing.T) {
	testCases := []struct {
		whenMethod   string
		whenPath     string
		expectPaths  []string
		expectParams []map[string]string
	}{
		{
			whenMethod:  http.MethodGet,
			whenPath:    "/",
			expectPaths: []string{"/"},
		},
		{
			whenMethod:   http.MethodGet,
			whenPath:     "/*",
			expectPaths:  []string{"/", "/foo"},
			expectParams: []map[string]string{{"*": ""}, {"*": "foo"}},
		},
		{
			whenMethod:   http.MethodGet,
			whenPath:     "/users/:id",
			expectPaths:  []string{"/users/0"},
			expectParams: []map[string]string{{"id": "0"}},
		},
		{
			whenMethod:   http.MethodGet,
			whenPath:     "/a/:b/c",
			expectPaths:  []string{"/a/b/c"},
			expectParams: []map[string]string{{"b": "b"}},
		}, {
			whenMethod:   http.MethodGet,
			whenPath:     "/a/*/c",
			expectPaths:  []string{"/a/b/c"},
			expectParams: []map[string]string{{"*": "b/c"}},
		},
		{
			whenMethod:   http.MethodGet,
			whenPath:     "/:a/b/c",
			expectPaths:  []string{"/a/b/c"},
			expectParams: []map[string]string{{"a": "a"}},
		}, {
			whenMethod:   http.MethodGet,
			whenPath:     "/*/b/c",
			expectPaths:  []string{"/a/b/c"},
			expectParams: []map[string]string{{"*": "a/b/c"}},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s %s", tc.whenMethod, tc.whenPath), func(t *testing.T) {
			n := NewRouteNode()
			defer n.Close()

			expectPort := node.MultiPort(node.PortOut, 0)

			err := n.Add(http.MethodGet, tc.whenPath, expectPort)
			assert.NoError(t, err)

			for i, expectPath := range tc.expectPaths {
				var expectParam map[string]string
				if len(tc.expectParams) > i {
					expectParam = tc.expectParams[i]
				}

				port, param := n.Find(http.MethodGet, expectPath)
				assert.Equal(t, expectPort, port)
				assert.Equal(t, expectParam, param)
			}
		})
	}
}

func TestRouteNode_Find(t *testing.T) {
	n := NewRouteNode()
	defer n.Close()

	n.Add(http.MethodGet, "/a/:b/c", node.MultiPort(node.PortOut, 0))
	n.Add(http.MethodGet, "/a/c/d", node.MultiPort(node.PortOut, 1))
	n.Add(http.MethodGet, "/a/c/df", node.MultiPort(node.PortOut, 2))
	n.Add(http.MethodGet, "/:e/c/f", node.MultiPort(node.PortOut, 3))
	n.Add(http.MethodGet, "/*", node.MultiPort(node.PortOut, 4))

	testCases := []struct {
		whenMethod   string
		whenPath     string
		expectPort   string
		expectParams map[string]string
	}{
		{
			whenMethod:   http.MethodGet,
			whenPath:     "/a/b/c",
			expectPort:   node.MultiPort(node.PortOut, 0),
			expectParams: map[string]string{"b": "b"},
		},
		{
			whenMethod: http.MethodGet,
			whenPath:   "/a/c/d",
			expectPort: node.MultiPort(node.PortOut, 1),
		},
		{
			whenMethod: http.MethodGet,
			whenPath:   "/a/c/df",
			expectPort: node.MultiPort(node.PortOut, 2),
		},
		{
			whenMethod:   http.MethodGet,
			whenPath:     "/e/c/f",
			expectPort:   node.MultiPort(node.PortOut, 3),
			expectParams: map[string]string{"e": "e"},
		},
		{
			whenMethod:   http.MethodGet,
			whenPath:     "/g/h/i",
			expectPort:   node.MultiPort(node.PortOut, 4),
			expectParams: map[string]string{"*": "g/h/i"},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s %s", tc.whenMethod, tc.whenPath), func(t *testing.T) {
			port, params := n.Find(tc.whenMethod, tc.whenPath)
			assert.Equal(t, tc.expectPort, port)
			assert.Equal(t, tc.expectParams, params)
		})
	}
}

func TestRouteNode_SendAndReceive(t *testing.T) {
	n := NewRouteNode()
	defer n.Close()

	n.Add(http.MethodGet, "/a/:b/c", node.MultiPort(node.PortOut, 0))
	n.Add(http.MethodGet, "/a/c/d", node.MultiPort(node.PortOut, 1))
	n.Add(http.MethodGet, "/a/*", node.MultiPort(node.PortOut, 2))

	testCases := []struct {
		whenMethod   string
		whenPath     string
		expectPort   string
		expectParams map[string]string
		expectStatus int
	}{
		{
			whenMethod:   http.MethodGet,
			whenPath:     "/a/b/c",
			expectPort:   node.MultiPort(node.PortOut, 0),
			expectParams: map[string]string{"b": "b"},
		},
		{
			whenMethod: http.MethodGet,
			whenPath:   "/a/c/d",
			expectPort: node.MultiPort(node.PortOut, 1),
		},
		{
			whenMethod:   http.MethodGet,
			whenPath:     "/a/d/e",
			expectPort:   node.MultiPort(node.PortOut, 2),
			expectParams: map[string]string{"*": "d/e"},
		},
		{
			whenMethod: http.MethodGet,
			whenPath:   "/b/c/d",
			expectPort: node.PortErr,
		},
		{
			whenMethod: http.MethodPost,
			whenPath:   "/a/b/c",
			expectPort: node.PortErr,
		},
		{
			whenMethod: http.MethodOptions,
			whenPath:   "/a/b/c",
			expectPort: node.PortErr,
		},
	}

	in := port.New()
	inPort, _ := n.Port(node.PortIn)
	inPort.Link(in)

	outs := map[string]*port.Port{}
	for _, tc := range testCases {
		if _, ok := outs[tc.expectPort]; !ok {
			out := port.New()
			outPort, _ := n.Port(tc.expectPort)
			outPort.Link(out)
			outs[tc.expectPort] = out
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s %s", tc.whenMethod, tc.whenPath), func(t *testing.T) {
			out := outs[tc.expectPort]

			proc := process.New()
			defer proc.Exit(nil)

			inStream := in.Open(proc)
			outStream := out.Open(proc)

			inPayload := primitive.NewMap(
				primitive.NewString("method"), primitive.NewString(tc.whenMethod),
				primitive.NewString("path"), primitive.NewString(tc.whenPath),
			)
			inPck := packet.New(inPayload)

			inStream.Send(inPck)

			ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
			defer cancel()

			select {
			case outPck := <-outStream.Receive():
				params, _ := primitive.Pick[map[string]string](outPck.Payload(), "params")
				assert.Equal(t, tc.expectParams, params)
				outStream.Send(outPck)
			case <-ctx.Done():
				assert.Fail(t, ctx.Err().Error())
			}
		})
	}
}
