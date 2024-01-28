// This code is adapted from the "github.com/labstack/echo" project, specifically the file "router.go," which is licensed under the MIT License.
package network

import (
	"bytes"
	"net/http"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// RouteNode represents a node for routing based on HTTP method, path, and port.
type RouteNode struct {
	*node.OneToManyNode
	tree *route
	mu   sync.RWMutex
}

// RouteNodeSpec defines the specification for configuring a RouteNode.
type RouteNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Routes          []Route `map:"routes"`
}

// Route represents a routing configuration with a specific HTTP method, path, and port.
type Route struct {
	Method string `map:"method"`
	Path   string `map:"path"`
	Port   string `map:"port"`
}

type route struct {
	kind           routeKind
	prefix         string
	paramNames     []string
	ports          map[string]string
	parent         *route
	staticChildren []*route
	paramChild     *route
	anyChild       *route
}

type routeKind uint8

const KindRoute = "route"

const (
	staticKind routeKind = iota
	paramKind
	anyKind

	paramLabel = byte(':')
	anyLabel   = byte('*')
)

// NewRouteNodeCodec creates a new codec for RouteNodeSpec.
func NewRouteNodeCodec() scheme.Codec {
	return scheme.CodecWithType[*RouteNodeSpec](func(spec *RouteNodeSpec) (node.Node, error) {
		n := NewRouteNode()
		for _, route := range spec.Routes {
			if err := n.Add(route.Method, route.Path, route.Port); err != nil {
				_ = n.Close()
				return nil, err
			}
		}
		return n, nil
	})
}

// NewRouteNode creates a new RouteNode.
func NewRouteNode() *RouteNode {
	n := &RouteNode{tree: &route{}}
	n.OneToManyNode = node.NewOneToManyNode(n.action)
	return n
}

// Add adds a new route to the routing tree for the specified HTTP method, path, and port.
func (n *RouteNode) Add(method, path, port string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	_, ok := node.IndexOfMultiPort(node.PortOut, port)
	if !ok {
		return errors.WithStack(node.ErrUnsupportedPort)
	}

	if path == "" {
		path = "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}

	paramNames := []string{}

	for i, lcpIndex := 0, len(path); i < lcpIndex; i++ {
		if path[i] == ':' {
			if i > 0 && path[i-1] == '\\' {
				path = path[:i-1] + path[i:]
				i--
				lcpIndex--
				continue
			}
			j := i + 1

			n.insert(method, path[:i], staticKind, nil, "")
			for ; i < lcpIndex && path[i] != '/'; i++ {
			}

			paramNames = append(paramNames, path[j:i])
			path = path[:j] + path[i:]
			i, lcpIndex = j, len(path)

			if i == lcpIndex {
				n.insert(method, path[:i], paramKind, paramNames, port)
			} else {
				n.insert(method, path[:i], paramKind, nil, "")
			}
		} else if path[i] == '*' {
			n.insert(method, path[:i], staticKind, nil, "")
			paramNames = append(paramNames, "*")
			n.insert(method, path[:i+1], anyKind, paramNames, port)
		}
	}

	n.insert(method, path, staticKind, paramNames, port)

	return nil
}

// Find searches for a matching route based on the provided HTTP method and path.
func (n *RouteNode) Find(method, path string) (string, map[string]string) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	route, paramValues := n.find(method, path)
	if route == nil {
		return "", nil
	}
	port := route.findPort(method)
	if port == "" {
		return "", nil
	}

	var params map[string]string
	if len(route.paramNames) > 0 {
		params = make(map[string]string, len(route.paramNames))
		for i, name := range route.paramNames {
			params[name] = paramValues[i]
		}
	}

	return port, params
}

func (n *RouteNode) action(proc *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inPayload, ok := inPck.Payload().(*primitive.Map)
	if !ok {
		return nil, nil
	}

	method, _ := primitive.Pick[string](inPayload, "method")
	path, _ := primitive.Pick[string](inPayload, "path")

	route, paramValues := n.find(method, path)
	if route == nil {
		outPayload, _ := primitive.MarshalBinary(PayloadNotFound)
		return nil, packet.New(outPayload)
	}

	port := route.findPort(method)
	if port == "" {
		var res HTTPPayload
		if method == http.MethodOptions {
			res = NewHTTPPayload(http.StatusNoContent, nil)
		} else {
			res = NewHTTPPayload(http.StatusMethodNotAllowed)
		}
		res.Header.Set(HeaderAllow, route.allowHeader())
		outPayload, _ := primitive.MarshalBinary(res)
		return nil, packet.New(outPayload)
	}

	params := make([]primitive.Value, 0, len(paramValues)*2)
	for i, name := range route.paramNames {
		params = append(params, primitive.NewString(name), primitive.NewString(paramValues[i]))
	}

	outPayload := inPayload.Set(primitive.NewString("params"), primitive.NewMap(params...))
	outPck := packet.New(outPayload)

	i, _ := node.IndexOfMultiPort(node.PortOut, port)

	outPcks := make([]*packet.Packet, i+1)
	outPcks[i] = outPck

	return outPcks, nil
}

func (n *RouteNode) insert(method, path string, kind routeKind, paramNames []string, port string) {
	cur := n.tree
	search := path

	// LCP - Longest Common Prefix (https://en.wikipedia.org/wiki/LCP_array)
	for {
		searchLen := len(search)
		prefixLen := len(cur.prefix)
		lcpLen := 0

		max := prefixLen
		if searchLen < max {
			max = searchLen
		}
		for ; lcpLen < max && search[lcpLen] == cur.prefix[lcpLen]; lcpLen++ {
		}

		if lcpLen == 0 {
			cur.prefix = search
			if port != "" {
				cur.kind = kind
				cur.paramNames = paramNames
				cur.addMethod(method, port)
			}
		} else if lcpLen < prefixLen {
			r := &route{
				kind:           cur.kind,
				prefix:         cur.prefix[lcpLen:],
				ports:          cur.ports,
				parent:         cur,
				staticChildren: cur.staticChildren,
				paramChild:     cur.paramChild,
				anyChild:       cur.anyChild,
			}

			for _, child := range cur.staticChildren {
				child.parent = r
			}
			if cur.paramChild != nil {
				cur.paramChild.parent = r
			}
			if cur.anyChild != nil {
				cur.anyChild.parent = r
			}

			cur.kind = staticKind
			cur.prefix = cur.prefix[:lcpLen]
			cur.staticChildren = nil
			cur.ports = map[string]string{}
			cur.paramChild = nil
			cur.anyChild = nil

			cur.addStaticChild(r)

			if lcpLen == searchLen {
				cur.kind = kind
				if port != "" {
					cur.paramNames = paramNames
					cur.addMethod(method, port)
				}
			} else {
				r := &route{
					kind:   kind,
					prefix: search[lcpLen:],
					ports:  map[string]string{},
					parent: cur,
				}

				if port != "" {
					r.paramNames = paramNames
					r.addMethod(method, port)
				}
				cur.addStaticChild(r)
			}
		} else if lcpLen < searchLen {
			search = search[lcpLen:]
			c := cur.findChild(search[0])
			if c != nil {
				cur = c
				continue
			}
			r := &route{
				kind:   kind,
				prefix: search,
				ports:  map[string]string{},
				parent: cur,
			}
			if port != "" {
				r.paramNames = paramNames
				r.addMethod(method, port)
			}

			switch kind {
			case staticKind:
				cur.addStaticChild(r)
			case paramKind:
				cur.paramChild = r
			case anyKind:
				cur.anyChild = r
			}
		} else if port != "" {
			cur.paramNames = paramNames
			cur.addMethod(method, port)
		}
		return
	}
}

func (n *RouteNode) find(method, path string) (*route, []string) {
	bestMatchedRoute := n.tree

	var (
		prevBestMatchedRoute *route
		search               = path
		searchIndex          = 0
		paramIndex           int
		paramValues          = []string{}
	)

	backtrackToNextRouteKind := func(fromKind routeKind) (nextRouteKind routeKind, valid bool) {
		prev := bestMatchedRoute
		bestMatchedRoute = prev.parent
		valid = bestMatchedRoute != nil

		if prev.kind == anyKind {
			nextRouteKind = staticKind
		} else {
			nextRouteKind = prev.kind + 1
		}

		if fromKind == staticKind {
			return
		}

		if prev.kind == staticKind {
			searchIndex -= len(prev.prefix)
		} else {
			paramIndex--
			searchIndex -= len(paramValues[paramIndex])
			paramValues[paramIndex] = ""
		}
		search = path[searchIndex:]
		return
	}

	for {
		prefixLen := 0
		lcpLen := 0

		if bestMatchedRoute.kind == staticKind {
			searchLen := len(search)
			prefixLen = len(bestMatchedRoute.prefix)

			max := prefixLen
			if searchLen < max {
				max = searchLen
			}
			for ; lcpLen < max && search[lcpLen] == bestMatchedRoute.prefix[lcpLen]; lcpLen++ {
			}
		}

		if lcpLen != prefixLen {
			if rk, ok := backtrackToNextRouteKind(staticKind); !ok {
				return nil, nil
			} else if rk == paramKind {
				goto Param
			} else {
				break
			}
		}

		search = search[lcpLen:]
		searchIndex = searchIndex + lcpLen

		if search == "" {
			if bestMatchedRoute.hasPort() {
				if prevBestMatchedRoute == nil {
					prevBestMatchedRoute = bestMatchedRoute
				}
				if port := bestMatchedRoute.findPort(method); port != "" {
					break
				}
			}
		}

		if search != "" {
			if child := bestMatchedRoute.findChild(search[0]); child != nil {
				bestMatchedRoute = child
				continue
			}
		}

	Param:
		if child := bestMatchedRoute.paramChild; search != "" && child != nil {
			bestMatchedRoute = child
			i := 0
			l := len(search)
			if !bestMatchedRoute.hasChild() {
				i = l
			} else {
				for ; i < l && search[i] != '/'; i++ {
				}
			}

			if len(paramValues) <= paramIndex {
				paramValues = append(paramValues, search[:i])
			} else {
				paramValues[paramIndex] = search[:i]
			}
			paramIndex++
			search = search[i:]
			searchIndex = searchIndex + i
			continue
		}

	Any:
		if child := bestMatchedRoute.anyChild; child != nil {
			bestMatchedRoute = child
			if len(paramValues) < len(bestMatchedRoute.paramNames) {
				paramValues = append(paramValues, search)
			} else {
				paramValues[len(bestMatchedRoute.paramNames)-1] = search
			}
			paramIndex++

			searchIndex += len(search)
			search = ""

			if port := bestMatchedRoute.findPort(method); port != "" {
				break
			}
			if prevBestMatchedRoute == nil {
				prevBestMatchedRoute = bestMatchedRoute
			}
		}

		if rk, ok := backtrackToNextRouteKind(anyKind); !ok {
			break
		} else if rk == paramKind {
			goto Param
		} else if rk == anyKind {
			goto Any
		} else {
			break
		}
	}

	if bestMatchedRoute == nil && prevBestMatchedRoute == nil {
		return nil, nil
	} else if bestMatchedRoute != nil {
		return bestMatchedRoute, paramValues
	} else {
		return prevBestMatchedRoute, nil
	}
}

func (r *route) addStaticChild(child *route) {
	r.staticChildren = append(r.staticChildren, child)
}

func (r *route) addMethod(method, port string) {
	if r.ports == nil {
		r.ports = make(map[string]string)
	}
	if port == "" {
		delete(r.ports, method)
	} else {
		r.ports[method] = port
	}
}

func (r *route) findChild(label byte) *route {
	for _, c := range r.staticChildren {
		if c.label() == label {
			return c
		}
	}
	if label == paramLabel {
		return r.paramChild
	}
	if label == anyLabel {
		return r.anyChild
	}
	return nil
}

func (r *route) hasChild() bool {
	return len(r.staticChildren) > 0 || r.anyChild != nil || r.paramChild != nil
}

func (r *route) findPort(method string) string {
	return r.ports[method]
}

func (r *route) hasPort() bool {
	return len(r.ports) > 0
}

func (r *route) allowHeader() string {
	buf := new(bytes.Buffer)
	buf.WriteString(http.MethodOptions)

	for method := range r.ports {
		buf.WriteString(", ")
		buf.WriteString(method)
	}
	return buf.String()
}

func (r *route) label() byte {
	return r.prefix[0]
}
