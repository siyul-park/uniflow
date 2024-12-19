package node

// nopCloserNode wraps a Node and makes Close a no-op.
type nopCloserNode struct {
	Node
}

// NoCloser returns a Node with a no-op Close method.
func NoCloser(node Node) Node {
	return &nopCloserNode{Node: node}
}

// Close does nothing and always returns nil.
func (*nopCloserNode) Close() error {
	return nil
}
