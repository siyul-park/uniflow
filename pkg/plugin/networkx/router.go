package networkx

import (
	"bytes"
	"net/http"
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// RouterNodeConfig holds the configuration for RouterNode.
type RouterNodeConfig struct {
	ID ulid.ULID
}

// RouterNode represents a router node that handles routing based on HTTP methods, paths, and ports.
type RouterNode struct {
	*node.OneToManyNode
	tree *route
	mu   sync.RWMutex
}

// RouterSpec defines the specification for the router node.
type RouterSpec struct {
	scheme.SpecMeta `map:",inline"`
	Routes          []RouteInfo `map:"routes"`
}

// RouteInfo holds information about an individual route.
type RouteInfo struct {
	Method string `map:"method"`
	Path   string `map:"path"`
	Port   string `map:"port"`
}

type route struct {
	kind           routeKind
	prefix         string
	parent         *route
	staticChildren []*route
	paramChild     *route
	anyChild       *route
	paramNames     []string
	methods        map[string]string
}

type routeKind uint8

// KindRouter is the kind identifier for RouterNode.
const KindRouter = "router"

const (
	KeyMethod = "method"
	KeyPath   = "path"
	KeyParams = "params"
)

const (
	staticKind routeKind = iota
	paramKind
	anyKind

	paramLabel = byte(':')
	anyLabel   = byte('*')
)

var _ node.Node = (*RouterNode)(nil)
var _ scheme.Spec = (*RouterSpec)(nil)

// NewRouterNode creates a new instance of RouterNode with the given configuration.
func NewRouterNode(config RouterNodeConfig) *RouterNode {
	id := config.ID

	n := &RouterNode{
		tree: &route{
			methods: map[string]string{},
		},
	}
	n.OneToManyNode = node.NewOneToManyNode(node.OneToManyNodeConfig{
		ID:     id,
		Action: n.action,
	})

	return n
}

// Add adds a new route to the router based on the provided HTTP method, path, and port.
func (n *RouterNode) Add(method, path, port string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if path == "" {
		path = "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}

	var paramNames []string

	for i, lcpIndex := 0, len(path); i < lcpIndex; i++ {
		if path[i] == paramLabel {
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
		} else if path[i] == anyLabel {
			n.insert(method, path[:i], staticKind, nil, "")
			paramNames = append(paramNames, "*")
			n.insert(method, path[:i+1], anyKind, paramNames, port)
		}
	}

	n.insert(method, path, staticKind, paramNames, port)
}

// Close resets the router's tree when closing the node.
func (n *RouterNode) Close() error {
	n.tree = &route{
		methods: map[string]string{},
	}
	return n.OneToManyNode.Close()
}

func (n *RouterNode) action(proc *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inPayload, ok := inPck.Payload().(*primitive.Map)
	if !ok {
		return nil, packet.WithError(packet.ErrInvalidPacket, inPck)
	}
	method, ok := primitive.Pick[string](inPayload, KeyMethod)
	if !ok {
		return nil, packet.WithError(packet.ErrInvalidPacket, inPck)
	}
	path, ok := primitive.Pick[string](inPayload, KeyPath)
	if !ok {
		return nil, packet.WithError(packet.ErrInvalidPacket, inPck)
	}

	pre, cur, values := n.find(method, path)

	if cur != nil {
		p := cur.methods[method]
		var paramPairs []primitive.Value
		for i, v := range values {
			paramPairs = append(paramPairs, primitive.NewString(cur.paramNames[i]))
			paramPairs = append(paramPairs, primitive.NewString(v))
		}

		if i, ok := port.GetIndex(node.PortOut, p); ok {
			outPayload := inPayload.Set(primitive.NewString(KeyParams), primitive.NewMap(paramPairs...))
			outPck := packet.New(outPayload)
			outPcks := make([]*packet.Packet, i+1)
			outPcks[i] = outPck

			return outPcks, nil
		}
	} else if pre != nil {
		buf := new(bytes.Buffer)
		buf.WriteString(http.MethodOptions)
		for k := range pre.methods {
			if k == http.MethodOptions {
				continue
			}
			buf.WriteString(", ")
			buf.WriteString(k)
		}

		header := http.Header(map[string][]string{
			HeaderAllow: {buf.String()},
		})

		if method == http.MethodOptions {
			errPayload, _ := primitive.MarshalText(HTTPPayload{
				Header: header,
				Status: http.StatusNoContent,
			})
			return nil, packet.New(errPayload)
		} else {
			errPayload, _ := primitive.MarshalText(HTTPPayload{
				Header: header,
				Body:   primitive.NewString(http.StatusText(http.StatusMethodNotAllowed)),
				Status: http.StatusMethodNotAllowed,
			})
			return nil, packet.New(errPayload)
		}
	}

	errPayload, _ := primitive.MarshalText(NotFound)
	return nil, packet.New(errPayload)
}

func (n *RouterNode) insert(method, path string, kind routeKind, paramNames []string, port string) {
	currentRoute := n.tree
	search := path

	for {
		searchLen := len(search)
		prefixLen := len(currentRoute.prefix)
		lcpLen := 0

		// LCP - Longest Common Prefix (https://en.wikipedia.org/wiki/LCP_array)
		max := prefixLen
		if searchLen < max {
			max = searchLen
		}
		for ; lcpLen < max && search[lcpLen] == currentRoute.prefix[lcpLen]; lcpLen++ {
		}

		if lcpLen == 0 {
			// At root node
			currentRoute.prefix = search
			if port != "" {
				currentRoute.kind = kind
				currentRoute.paramNames = paramNames
				currentRoute.methods[method] = port
			}
		} else if lcpLen < prefixLen {
			r := &route{
				kind:           currentRoute.kind,
				prefix:         currentRoute.prefix[lcpLen:],
				parent:         currentRoute,
				staticChildren: currentRoute.staticChildren,
				paramChild:     currentRoute.paramChild,
				anyChild:       currentRoute.anyChild,
				paramNames:     currentRoute.paramNames,
				methods:        currentRoute.methods,
			}
			for _, child := range currentRoute.staticChildren {
				child.parent = r
			}
			if currentRoute.paramChild != nil {
				currentRoute.paramChild.parent = r
			}
			if currentRoute.anyChild != nil {
				currentRoute.anyChild.parent = r
			}

			// Reset parent node
			currentRoute.kind = staticKind
			currentRoute.prefix = currentRoute.prefix[:lcpLen]
			currentRoute.staticChildren = nil
			currentRoute.paramNames = nil
			currentRoute.paramChild = nil
			currentRoute.anyChild = nil
			currentRoute.methods = map[string]string{}

			// Only Static children could reach here
			currentRoute.addStaticChild(r)

			if lcpLen == searchLen {
				// At parent node
				currentRoute.kind = kind
				if port != "" {
					currentRoute.paramNames = paramNames
					currentRoute.methods[method] = port
				}
			} else {
				// Create child node
				r = &route{
					kind:    kind,
					prefix:  search[lcpLen:],
					parent:  currentRoute,
					methods: map[string]string{},
				}
				if port != "" {
					r.paramNames = paramNames
					r.methods[method] = port
				}
				// Only Static children could reach here
				currentRoute.addStaticChild(r)
			}
		} else if lcpLen < searchLen {
			search = search[lcpLen:]
			c := currentRoute.findChildWithLabel(search[0])
			if c != nil {
				// Go deeper
				currentRoute = c
				continue
			}
			// Create child node
			r := &route{
				kind:    kind,
				prefix:  search,
				parent:  currentRoute,
				methods: map[string]string{},
			}
			if port != "" {
				r.paramNames = paramNames
				r.methods[method] = port
			}

			switch kind {
			case staticKind:
				currentRoute.addStaticChild(r)
			case paramKind:
				currentRoute.paramChild = r
			case anyKind:
				currentRoute.anyChild = r
			}
		} else {
			// Node already exists
			if port != "" {
				currentRoute.paramNames = paramNames
				currentRoute.methods[method] = port
			}
		}
		return
	}
}

func (n *RouterNode) find(method, path string) (*route, *route, []string) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	currentRoute := n.tree

	var (
		bestMatchRoute         *route
		previousBestMatchRoute *route
		search                 = path
		searchIndex            = 0
		paramValues            []string
	)

	backtrackToNextRouteKind := func(fromKind routeKind) (nextNodeKind routeKind, valid bool) {
		previous := currentRoute
		currentRoute = previous.parent
		valid = currentRoute != nil

		// Next node type by priority
		if previous.kind == anyKind {
			nextNodeKind = staticKind
		} else {
			nextNodeKind = previous.kind + 1
		}

		if fromKind == staticKind {
			// when backtracking is done from static basisKind block we did not change search so nothing to restore
			return
		}

		// restore search to value it was before we move to current node we are backtracking from.
		if previous.kind == staticKind {
			searchIndex -= len(previous.prefix)
		} else if len(paramValues) > 0 {
			searchIndex -= len(paramValues[len(paramValues)-1])
			paramValues = paramValues[:len(paramValues)-1]
		}
		search = path[searchIndex:]
		return
	}

	for {
		prefixLen := 0
		lcpLen := 0

		if currentRoute.kind == staticKind {
			searchLen := len(search)
			prefixLen = len(currentRoute.prefix)

			// LCP - Longest Common Prefix (https://en.wikipedia.org/wiki/LCP_array)
			max := prefixLen
			if searchLen < max {
				max = searchLen
			}
			for ; lcpLen < max && search[lcpLen] == currentRoute.prefix[lcpLen]; lcpLen++ {
			}
		}

		if lcpLen != prefixLen {
			// No matching prefix, let's backtrack to the first possible alternative node of the decision path
			rk, ok := backtrackToNextRouteKind(staticKind)
			if !ok {
				return nil, nil, nil
			} else if rk == paramKind {
				goto Param
			} else {
				// Not found (this should never be possible for static node we are looking currently)
				break
			}
		}

		// The full prefix has matched, remove the prefix from the remaining search
		search = search[lcpLen:]
		searchIndex = searchIndex + lcpLen

		// Finish routing if is no request path remaining to search
		if search == "" {
			if currentRoute.hasPort() {
				if previousBestMatchRoute == nil {
					previousBestMatchRoute = currentRoute
				}
				if _, ok := currentRoute.methods[method]; ok {
					bestMatchRoute = currentRoute
					break
				}
			}
		}

		// Static node
		if search != "" {
			if child := currentRoute.findStaticChild(search[0]); child != nil {
				currentRoute = child
				continue
			}
		}

	Param:
		// Param node
		if child := currentRoute.paramChild; search != "" && child != nil {
			currentRoute = child
			i := 0
			l := len(search)
			if currentRoute.isLeaf() {
				// when param node does not have any children (path param is last piece of route path) then param node should
				// act similarly to any node - consider all remaining search as match
				i = l
			} else {
				for ; i < l && search[i] != '/'; i++ {
				}
			}

			paramValues = append(paramValues, search[:i])
			search = search[i:]
			searchIndex = searchIndex + i
			continue
		}

	Any:
		// Any node
		if child := currentRoute.anyChild; child != nil {
			// If any node is found, use remaining path for paramValues
			currentRoute = child
			paramValues = append(paramValues, search)

			// update indexes/search in case we need to backtrack when no handler match is found
			searchIndex += +len(search)
			search = ""

			if _, ok := currentRoute.methods[method]; ok {
				bestMatchRoute = currentRoute
				break
			}
			if previousBestMatchRoute == nil {
				previousBestMatchRoute = currentRoute
			}
		}

		// Let's backtrack to the first possible alternative node of the decision path
		rk, ok := backtrackToNextRouteKind(anyKind)
		if !ok {
			break // No other possibilities on the decision path
		} else if rk == paramKind {
			goto Param
		} else if rk == anyKind {
			goto Any
		} else {
			// Not found
			break
		}
	}

	return previousBestMatchRoute, bestMatchRoute, paramValues
}

func (r *route) addStaticChild(c *route) {
	r.staticChildren = append(r.staticChildren, c)
}

func (r *route) findChildWithLabel(l byte) *route {
	if c := r.findStaticChild(l); c != nil {
		return c
	}
	if l == paramLabel {
		return r.paramChild
	}
	if l == anyLabel {
		return r.anyChild
	}
	return nil
}

func (r *route) findStaticChild(l byte) *route {
	for _, c := range r.staticChildren {
		if c.label() == l {
			return c
		}
	}
	return nil
}

func (r *route) isLeaf() bool {
	return len(r.staticChildren) == 0 && r.paramChild == nil && r.anyChild == nil
}

func (r *route) hasPort() bool {
	return len(r.methods) > 0
}

func (r *route) label() byte {
	return r.prefix[0]
}
