package node

// Proxy is an interface for unwrapping a Node.
type Proxy interface {
	// Unwrap returns the underlying Node.
	Unwrap() Node
}

type noCloseNode struct {
	Node
}

var _ Proxy = (*noCloseNode)(nil)

// Unwrap recursively unwraps a Node if it implements Proxy.
func Unwrap(n Node) Node {
	proxy, ok := n.(Proxy)
	if !ok {
		return nil
	}
	return proxy.Unwrap()
}

// As attempts to cast the source Node to the target type T.
func As[T any](source Node, target *T) bool {
	for source != nil {
		if s, ok := source.(T); ok {
			*target = s
			return true
		}
		source = Unwrap(source)
	}
	return false
}

// NoCloser returns a Node with a no-op Exit method.
func NoCloser(node Node) Node {
	return &noCloseNode{Node: node}
}

// Unwrap returns the underlying Node in a noCloseNode.
func (n *noCloseNode) Unwrap() Node {
	return n.Node
}

// Close does nothing and always returns nil.
func (*noCloseNode) Close() error {
	return nil
}
