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
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestRouteNodeCodec_Decode(t *testing.T) {
	codec := NewRouteNodeCodec()

	spec := &RouteNodeSpec{
		Routes: []Route{
			{
				Method: http.MethodGet,
				Path:   "/",
				Port:   node.PortWithIndex(node.PortOut, 0),
			},
		},
	}

	n, err := codec.Compile(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

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

			expectPort := node.PortWithIndex(node.PortOut, 0)

			n.Add(http.MethodGet, tc.whenPath, expectPort)

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

	n.Add(http.MethodGet, "/a/:b/c", node.PortWithIndex(node.PortOut, 0))
	n.Add(http.MethodGet, "/a/c/d", node.PortWithIndex(node.PortOut, 1))
	n.Add(http.MethodGet, "/a/c/df", node.PortWithIndex(node.PortOut, 2))
	n.Add(http.MethodGet, "/:e/c/f", node.PortWithIndex(node.PortOut, 3))
	n.Add(http.MethodGet, "/*", node.PortWithIndex(node.PortOut, 4))

	testCases := []struct {
		whenMethod   string
		whenPath     string
		expectPort   string
		expectParams map[string]string
	}{
		{
			whenMethod:   http.MethodGet,
			whenPath:     "/a/b/c",
			expectPort:   node.PortWithIndex(node.PortOut, 0),
			expectParams: map[string]string{"b": "b"},
		},
		{
			whenMethod: http.MethodGet,
			whenPath:   "/a/c/d",
			expectPort: node.PortWithIndex(node.PortOut, 1),
		},
		{
			whenMethod: http.MethodGet,
			whenPath:   "/a/c/df",
			expectPort: node.PortWithIndex(node.PortOut, 2),
		},
		{
			whenMethod:   http.MethodGet,
			whenPath:     "/e/c/f",
			expectPort:   node.PortWithIndex(node.PortOut, 3),
			expectParams: map[string]string{"e": "e"},
		},
		{
			whenMethod:   http.MethodGet,
			whenPath:     "/g/h/i",
			expectPort:   node.PortWithIndex(node.PortOut, 4),
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

	n.Add(http.MethodGet, "/a/:b/c", node.PortWithIndex(node.PortOut, 0))
	n.Add(http.MethodGet, "/a/c/d", node.PortWithIndex(node.PortOut, 1))
	n.Add(http.MethodGet, "/a/*", node.PortWithIndex(node.PortOut, 2))

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
			expectPort:   node.PortWithIndex(node.PortOut, 0),
			expectParams: map[string]string{"b": "b"},
		},
		{
			whenMethod:   http.MethodGet,
			whenPath:     "/a/c/d",
			expectPort:   node.PortWithIndex(node.PortOut, 1),
			expectParams: map[string]string{},
		},
		{
			whenMethod:   http.MethodGet,
			whenPath:     "/a/d/e",
			expectPort:   node.PortWithIndex(node.PortOut, 2),
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

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	outs := map[string]*port.InPort{}
	for _, tc := range testCases {
		if _, ok := outs[tc.expectPort]; !ok {
			out := port.NewIn()
			n.Out(tc.expectPort).Link(out)
			outs[tc.expectPort] = out
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s %s", tc.whenMethod, tc.whenPath), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
			defer cancel()

			out := outs[tc.expectPort]

			proc := process.New()
			defer proc.Exit(nil)

			inWriter := in.Open(proc)
			outReader := out.Open(proc)

			inPayload := types.NewMap(
				types.NewString("method"), types.NewString(tc.whenMethod),
				types.NewString("path"), types.NewString(tc.whenPath),
			)
			inPck := packet.New(inPayload)

			inWriter.Write(inPck)

			select {
			case outPck := <-outReader.Read():
				params, _ := types.Pick[map[string]string](outPck.Payload(), "params")
				assert.Equal(t, tc.expectParams, params)
				outReader.Receive(outPck)
			case <-ctx.Done():
				assert.Fail(t, ctx.Err().Error())
			}

			select {
			case backPck := <-inWriter.Receive():
				assert.NotNil(t, backPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}
		})
	}
}


func BenchmarkRouteNode_SendAndReceive(b *testing.B) {
	n := NewRouteNode()
	defer n.Close()

	n.Add(http.MethodGet, "/a/b/c", node.PortWithIndex(node.PortOut, 0))

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	out0 := port.NewIn()
	n.Out(node.PortWithIndex(node.PortOut, 0)).Link(out0)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	outReader0 := out0.Open(proc)

	inPayload := types.NewMap(
		types.NewString("method"), types.NewString(http.MethodGet),
		types.NewString("path"), types.NewString("/a/b/c"),
	)
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		outPck := <-outReader0.Read()
		outReader0.Receive(outPck)
		<-inWriter.Receive()
	}
}
