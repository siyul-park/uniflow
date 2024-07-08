package scheme

// Registrar is an interface for registering types with a Scheme.
type Register interface {
	AddToScheme(*Scheme) error
}

// RegisterFunc is a function type that registers types with a Scheme.
type RegisterFunc func(*Scheme) error

var _ Register = (RegisterFunc)(nil)

// AddToScheme calls the RegisterFunc to register types with the Scheme.
func (f RegisterFunc) AddToScheme(s *Scheme) error {
	return f(s)
}
