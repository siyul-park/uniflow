package hook

// Register defines an interface for registering types with a Hook.
type Register interface {
	// AddToHooks adds types to the given Hook.
	AddToHook(*Hook) error
}

type register struct {
	addToHooks func(*Hook) error
}

var _ Register = (*register)(nil)

// RegisterFunc creates a new Register from the provided function.
func RegisterFunc(addToHooks func(*Hook) error) Register {
	return &register{addToHooks: addToHooks}
}

func (r *register) AddToHook(s *Hook) error {
	return r.addToHooks(s)
}
