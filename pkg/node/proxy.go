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
	if proxy, ok := n.(Proxy); ok {
		return Unwrap(proxy.Unwrap())
	}
	return n
}

// As attempts to cast the source Node to the target type T.
func As[T any](source Node, target *T) bool {
	for {
		if s, ok := source.(T); ok {
			*target = s
			return true
		}
		if proxy, ok := source.(Proxy); ok {
			source = proxy.Unwrap()
		} else {
			return false
		}
	}
}

// NoCloser returns a Node with a no-op Close method.
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
