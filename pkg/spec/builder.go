package spec

// Builder is a collection of functions to construct a new Scheme.
type Builder []func(*Scheme) error

// NewBuilder creates a new SchemeBuilder with the provided functions.
func NewBuilder(funcs ...func(*Scheme) error) Builder {
	return Builder(funcs)
}

// AddToScheme integrates all registered types into the given Scheme.
func (b *Builder) AddToScheme(s *Scheme) error {
	for _, f := range *b {
		if err := f(s); err != nil {
			return err
		}
	}
	return nil
}

// Register incorporates one or more functions to register Spec types.
func (b *Builder) Register(funcs ...func(*Scheme) error) {
	*b = append(*b, funcs...)
}

// Build yields a new Scheme containing the registered types.
func (b *Builder) Build() (*Scheme, error) {
	s := NewScheme()
	if err := b.AddToScheme(s); err != nil {
		return nil, err
	}
	return s, nil
}
