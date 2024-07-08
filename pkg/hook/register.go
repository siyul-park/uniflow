package hook

// Registrar is an interface for registering types with a Hook.
type Register interface {
	AddToHooks(*Hook) error
}

// RegisterFunc is a function type that registers types with a Hook.
type RegisterFunc func(*Hook) error

var _ Register = (RegisterFunc)(nil)

// AddToScheme calls the RegisterFunc to register types with the Hook.
func (f RegisterFunc) AddToHooks(h *Hook) error {
	return f(h)
}
