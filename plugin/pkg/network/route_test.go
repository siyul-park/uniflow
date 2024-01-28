package network

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/siyul-park/uniflow/pkg/node"
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
			whenMethod:  http.MethodGet,
			whenPath:    "/*",
			expectPaths: []string{"/", "/foo"},
			expectParams: []map[string]string{
				{
					"*": "",
				},
				{
					"*": "foo",
				},
			},
		},
		{
			whenMethod:  http.MethodGet,
			whenPath:    "/users/:id",
			expectPaths: []string{"/users/0"},
			expectParams: []map[string]string{
				{
					"id": "0",
				},
			},
		},
		{
			whenMethod:  http.MethodGet,
			whenPath:    "/a/:b/c",
			expectPaths: []string{"/a/b/c"},
			expectParams: []map[string]string{
				{
					"b": "b",
				},
			},
		}, {
			whenMethod:  http.MethodGet,
			whenPath:    "/a/*/c",
			expectPaths: []string{"/a/b/c"},
			expectParams: []map[string]string{
				{
					"*": "b/c", // NOTE: `b` would be probably more suitable
				},
			},
		},
		{
			whenMethod:  http.MethodGet,
			whenPath:    "/:a/b/c",
			expectPaths: []string{"/a/b/c"},
			expectParams: []map[string]string{
				{
					"a": "a",
				},
			},
		}, {
			whenMethod:  http.MethodGet,
			whenPath:    "/*/b/c",
			expectPaths: []string{"/a/b/c"},
			expectParams: []map[string]string{
				{
					"*": "a/b/c", // NOTE: `a` would be probably more suitable
				},
			},
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
