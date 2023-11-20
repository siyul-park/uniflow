package node

type (
	Wrapper interface {
		Wrap(Node) error
		Unwrap() Node
	}
)

// Unwrap unwraps all nested Wrapper.
func Unwrap(node Node) Node {
	for {
		if wrapper, ok := node.(Wrapper); ok {
			node = wrapper.Unwrap()
		} else {
			return node
		}
	}
}
