package scheme

// Builder is a collection of Register functions used to construct a new Scheme.
type Builder []Register

var _ Register = (*Builder)(nil)

// NewBuilder creates a new Builder with the provided Register functions.
func NewBuilder(registers ...Register) *Builder {
	b := &Builder{}
	b.Register(registers...)
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

// Register appends one or more Register functions to the Builder.
func (b *Builder) Register(registers ...Register) {
	*b = append(*b, registers...)
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
