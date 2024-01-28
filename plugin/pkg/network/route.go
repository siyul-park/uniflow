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
)

type RouteNode struct {
	*node.OneToManyNode
	tree *route
	mu   sync.RWMutex
}

type route struct {
	kind           routeKind
	prefix         string
	originalPath   string
	methods        map[string]*routeMethod
	parent         *route
	staticChildren []*route
	paramChild     *route
	anyChild       *route
}

type routeKind uint8

type routeMethod struct {
	originalPath string
	paramNames   []string
	port         string
}

const (
	staticKind routeKind = iota
	paramKind
	anyKind

	paramLabel = byte(':')
	anyLabel   = byte('*')
)

func NewRouteNode() *RouteNode {
	n := &RouteNode{
		tree: &route{
			methods: map[string]*routeMethod{},
		},
	}
	n.OneToManyNode = node.NewOneToManyNode(n.action)
	return n
}

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

	originalPath := path
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

			n.insert(method, path[:i], staticKind, routeMethod{})
			for ; i < lcpIndex && path[i] != '/'; i++ {
			}

			paramNames = append(paramNames, path[j:i])
			path = path[:j] + path[i:]
			i, lcpIndex = j, len(path)

			if i == lcpIndex {
				n.insert(method, path[:i], paramKind, routeMethod{originalPath, paramNames, port})
			} else {
				n.insert(method, path[:i], paramKind, routeMethod{})
			}
		} else if path[i] == '*' {
			n.insert(method, path[:i], staticKind, routeMethod{})
			paramNames = append(paramNames, "*")
			n.insert(method, path[:i+1], anyKind, routeMethod{originalPath, paramNames, port})
		}
	}

	n.insert(method, path, staticKind, routeMethod{originalPath, paramNames, port})

	return nil
}

func (n *RouteNode) Find(method, path string) (string, map[string]string) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	route, paramValues := n.find(method, path)
	if route == nil {
		return "", nil
	}
	rmethod := route.findMethod(method)
	if rmethod == nil {
		return "", nil
	}

	var params map[string]string
	if len(rmethod.paramNames) > 0 {
		params = make(map[string]string, len(rmethod.paramNames))
		for i, name := range rmethod.paramNames {
			params[name] = paramValues[i]
		}
	}

	return rmethod.port, params
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

	rmethod := route.findMethod(method)
	if rmethod == nil {
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
	for i, name := range rmethod.paramNames {
		params = append(params, primitive.NewString(name), primitive.NewString(paramValues[i]))
	}

	outPayload := inPayload.Set(primitive.NewString("params"), primitive.NewMap(params...))
	outPck := packet.New(outPayload)

	i, _ := node.IndexOfMultiPort(node.PortOut, rmethod.port)

	outPcks := make([]*packet.Packet, i+1)
	outPcks[i] = outPck

	return outPcks, nil
}

func (n *RouteNode) insert(method, path string, rkind routeKind, rmethod routeMethod) {
	cur := n.tree
	search := path

	for {
		searchLen := len(search)
		prefixLen := len(cur.prefix)
		lcpLen := 0

		// LCP - Longest Common Prefix (https://en.wikipedia.org/wiki/LCP_array)
		max := prefixLen
		if searchLen < max {
			max = searchLen
		}
		for ; lcpLen < max && search[lcpLen] == cur.prefix[lcpLen]; lcpLen++ {
		}

		if lcpLen == 0 {
			// At root node
			cur.prefix = search
			if rmethod.port != "" {
				cur.kind = rkind
				cur.addMethod(method, &rmethod)
				cur.originalPath = rmethod.originalPath
			}
		} else if lcpLen < prefixLen {
			// Split node into two before we insert new node.
			// This happens when we are inserting path that is submatch of any existing inserted paths.
			// For example, we have node `/test` and now are about to insert `/te/*`. In that case
			// 1. overlapping part is `/te` that is used as parent node
			// 2. `st` is part from existing node that is not matching - it gets its own node (child to `/te`)
			// 3. `/*` is the new part we are about to insert (child to `/te`)
			r := &route{
				kind:           cur.kind,
				prefix:         cur.prefix[lcpLen:],
				originalPath:   cur.originalPath,
				methods:        cur.methods,
				parent:         cur,
				staticChildren: cur.staticChildren,
				paramChild:     cur.paramChild,
				anyChild:       cur.anyChild,
			}

			// Update parent path for all children to new node
			for _, child := range cur.staticChildren {
				child.parent = r
			}
			if cur.paramChild != nil {
				cur.paramChild.parent = r
			}
			if cur.anyChild != nil {
				cur.anyChild.parent = r
			}

			// Reset parent node
			cur.kind = staticKind
			cur.prefix = cur.prefix[:lcpLen]
			cur.staticChildren = nil
			cur.originalPath = ""
			cur.methods = map[string]*routeMethod{}
			cur.paramChild = nil
			cur.anyChild = nil

			// Only Static children could reach here
			cur.addStaticChild(r)

			if lcpLen == searchLen {
				// At parent node
				cur.kind = rkind
				if rmethod.port != "" {
					cur.addMethod(method, &rmethod)
					cur.originalPath = rmethod.originalPath
				}
			} else {
				// Create child node
				r := &route{
					kind:         rkind,
					prefix:       search[lcpLen:],
					originalPath: cur.originalPath,
					methods:      map[string]*routeMethod{},
					parent:       cur,
				}

				if rmethod.port != "" {
					r.addMethod(method, &rmethod)
					r.originalPath = rmethod.originalPath
				}
				// Only Static children could reach here
				cur.addStaticChild(r)
			}
		} else if lcpLen < searchLen {
			search = search[lcpLen:]
			c := cur.findChild(search[0])
			if c != nil {
				// Go deeper
				cur = c
				continue
			}
			// Create child node
			r := &route{
				kind:         rkind,
				prefix:       search,
				originalPath: rmethod.originalPath,
				methods:      map[string]*routeMethod{},
				parent:       cur,
			}
			if rmethod.port != "" {
				r.addMethod(method, &rmethod)
			}

			switch rkind {
			case staticKind:
				cur.addStaticChild(r)
			case paramKind:
				cur.paramChild = r
			case anyKind:
				cur.anyChild = r
			}
		} else if rmethod.port != "" { // Node already exists
			cur.addMethod(method, &rmethod)
			cur.originalPath = rmethod.originalPath
		}
		return
	}
}

func (n *RouteNode) find(method, path string) (*route, []string) {
	bestMatchedRoute := n.tree

	var (
		prevBestMatchedRoute *route
		// search stores the remaining path to check for match. By each iteration we move from start of path to end of the path
		// and search value gets shorter and shorter.
		search      = path
		searchIndex = 0
		paramIndex  int          // Param counter
		paramValues = []string{} // Use the internal slice so the interface can keep the illusion of a dynamic slice
	)

	// Backtracking is needed when a dead end (leaf node) is reached in the router tree.
	// To backtrack the current node will be changed to the parent node and the next kind for the
	// router logic will be returned based on fromKind or kind of the dead end node (static > param > any).
	// For example if there is no static node match we should check parent next sibling by kind (param).
	// Backtracking itself does not check if there is a next sibling, this is done by the router logic.
	backtrackToNextRouteKind := func(fromKind routeKind) (nextRouteKind routeKind, valid bool) {
		prev := bestMatchedRoute
		bestMatchedRoute = prev.parent
		valid = bestMatchedRoute != nil

		// Next node type by priority
		if prev.kind == anyKind {
			nextRouteKind = staticKind
		} else {
			nextRouteKind = prev.kind + 1
		}

		if fromKind == staticKind {
			// when backtracking is done from static kind block we did not change search so nothing to restore
			return
		}

		// restore search to value it was before we move to current node we are backtracking from.
		if prev.kind == staticKind {
			searchIndex -= len(prev.prefix)
		} else {
			paramIndex--
			// for param/any node.prefix value is always `:` so we can not deduce searchIndex from that and must use pValue
			// for that index as it would also contain part of path we cut off before moving into node we are backtracking from
			searchIndex -= len(paramValues[paramIndex])
			paramValues[paramIndex] = ""
		}
		search = path[searchIndex:]
		return
	}

	// Router tree is implemented by longest common prefix array (LCP array) https://en.wikipedia.org/wiki/LCP_array
	// Tree search is implemented as for loop where one loop iteration is divided into 3 separate blocks
	// Each of these blocks checks specific kind of node (static/param/any). Order of blocks reflex their priority in routing.
	// Search order/priority is: static > param > any.
	//
	// Note: backtracking in tree is implemented by replacing/switching currentNode to previous node
	// and hoping to (goto statement) next block by priority to check if it is the match.
	for {
		prefixLen := 0 // Prefix length
		lcpLen := 0    // LCP (longest common prefix) length

		if bestMatchedRoute.kind == staticKind {
			searchLen := len(search)
			prefixLen = len(bestMatchedRoute.prefix)

			// LCP - Longest Common Prefix (https://en.wikipedia.org/wiki/LCP_array)
			max := prefixLen
			if searchLen < max {
				max = searchLen
			}
			for ; lcpLen < max && search[lcpLen] == bestMatchedRoute.prefix[lcpLen]; lcpLen++ {
			}
		}

		if lcpLen != prefixLen {
			// No matching prefix, let's backtrack to the first possible alternative node of the decision path
			nk, ok := backtrackToNextRouteKind(staticKind)
			if !ok {
				return nil, nil // No other possibilities on the decision path, handler will be whatever context is reset to.
			} else if nk == paramKind {
				goto Param
				// NOTE: this case (backtracking from static node to previous any node) can not happen by current any matching logic. Any node is end of search currently
				//} else if nk == anyKind {
				//	goto Any
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
			// in case of node that is handler we have exact method type match or something for 405 to use
			if bestMatchedRoute.hasMethod() {
				// check if current node has handler registered for http method we are looking for. we store currentNode as
				// best matching in case we do no find no more routes matching this path+method
				if prevBestMatchedRoute == nil {
					prevBestMatchedRoute = bestMatchedRoute
				}
				if m := bestMatchedRoute.findMethod(method); m != nil {
					break
				}
			}
		}

		// Static node
		if search != "" {
			if child := bestMatchedRoute.findChild(search[0]); child != nil {
				bestMatchedRoute = child
				continue
			}
		}

	Param:
		// Param node
		if child := bestMatchedRoute.paramChild; search != "" && child != nil {
			bestMatchedRoute = child
			i := 0
			l := len(search)
			if !bestMatchedRoute.hasChild() {
				// when param node does not have any children (path param is last piece of route path) then param node should
				// act similarly to any node - consider all remaining search as match
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
		// Any node
		if child := bestMatchedRoute.anyChild; child != nil {
			// If any node is found, use remaining path for paramValues
			bestMatchedRoute = child
			if len(paramValues) < bestMatchedRoute.paramLen() {
				paramValues = append(paramValues, search)
			} else {
				paramValues[bestMatchedRoute.paramLen()-1] = search
			}
			paramIndex++

			// update indexes/search in case we need to backtrack when no handler match is found
			searchIndex += len(search)
			search = ""

			if h := bestMatchedRoute.findMethod(method); h != nil {
				break
			}
			// we store currentNode as best matching in case we do not find more routes matching this path+method. Needed for 405
			if prevBestMatchedRoute == nil {
				prevBestMatchedRoute = bestMatchedRoute
			}
		}

		// Let's backtrack to the first possible alternative node of the decision path
		nk, ok := backtrackToNextRouteKind(anyKind)
		if !ok {
			break // No other possibilities on the decision path
		} else if nk == paramKind {
			goto Param
		} else if nk == anyKind {
			goto Any
		} else {
			// Not found
			break
		}
	}

	if bestMatchedRoute == nil && prevBestMatchedRoute == nil {
		return nil, nil
	}

	if bestMatchedRoute != nil {
		return bestMatchedRoute, paramValues
	} else {
		return prevBestMatchedRoute, nil
	}
}

func (r *route) addStaticChild(child *route) {
	r.staticChildren = append(r.staticChildren, child)
}

func (r *route) addMethod(method string, rm *routeMethod) {
	if r.methods == nil {
		r.methods = make(map[string]*routeMethod)
	}
	if rm.port == "" {
		delete(r.methods, method)
	} else {
		r.methods[method] = rm
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

func (r *route) findMethod(method string) *routeMethod {
	return r.methods[method]
}

func (r *route) hasChild() bool {
	return len(r.staticChildren) > 0 || r.anyChild != nil || r.paramChild != nil
}

func (r *route) hasMethod() bool {
	return len(r.methods) > 0
}

func (r *route) allowHeader() string {
	buf := new(bytes.Buffer)
	buf.WriteString(http.MethodOptions)

	for method := range r.methods {
		buf.WriteString(", ")
		buf.WriteString(method)
	}
	return buf.String()
}

func (r *route) paramLen() int {
	count := 0
	for _, method := range r.methods {
		if len(method.paramNames) > count {
			count = len(method.paramNames)
		}
	}
	return count
}

func (r *route) label() byte {
	return r.prefix[0]
}
