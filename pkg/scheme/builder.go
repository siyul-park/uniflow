package scheme

// Builder is a collection of Register functions used to construct a new Scheme.
type Builder []Register

var _ Register = (*Builder)(nil)

// NewBuilder creates a new Builder with the provided Register functions.
func NewBuilder(registers ...Register) *Builder {
	b := &Builder{}
	for _, r := range registers {
		b.Register(r)
	}
	return b
}

// AddToScheme applies all Register functions in the Builder to the given Scheme.
func (b *Builder) AddToScheme(s *Scheme) error {
	for _, f := range *b {
		if err := f.AddToScheme(s); err != nil {
			return err
		}
	}
	return nil
}

// Register adds a Register function if not already present.
func (b *Builder) Register(register Register) bool {
	for _, r := range *b {
		if r == register {
			return false
		}
	}
	*b = append(*b, register)
	return true
}

// Unregister removes a Register function if present.
func (b *Builder) Unregister(register Register) bool {
	for i, r := range *b {
		if r == register {
			*b = append((*b)[:i], (*b)[i+1:]...)
			return true
		}
	}
	return false
}

// Len returns the number of registered hook functions.
func (b *Builder) Len() int {
	return len(*b)
}

// Build creates a new Scheme and applies all Register functions to it.
func (b *Builder) Build() (*Scheme, error) {
	s := New()
	if err := b.AddToScheme(s); err != nil {
		return nil, err
	}
	return s, nil
}
