package scheme

// Register defines an interface for registering types with a Scheme.
type Register interface {
	// AddToScheme adds types to the given Scheme.
	AddToScheme(*Scheme) error
}

type register struct {
	addToScheme func(*Scheme) error
}

var _ Register = (*register)(nil)

// RegisterFunc creates a new Register from the provided function.
func RegisterFunc(addToScheme func(*Scheme) error) Register {
	return &register{addToScheme: addToScheme}
}

func (r *register) AddToScheme(s *Scheme) error {
	return r.addToScheme(s)
}
