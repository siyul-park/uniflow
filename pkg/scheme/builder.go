package scheme

// Builder is a collection of functions to build a new Scheme.
type Builder []func(*Scheme) error

// NewBuilder creates a new SchemeBuilder with the provided functions.
func NewBuilder(funcs ...func(*Scheme) error) Builder {
	return Builder(funcs)
}

// AddToScheme adds all registered types to the given Scheme.
func (b *Builder) AddToScheme(s *Scheme) error {
	for _, f := range *b {
		if err := f(s); err != nil {
			return err
		}
	}
	return nil
}

// Register adds one or more functions to register Spec types.
func (b *Builder) Register(funcs ...func(*Scheme) error) {
	*b = append(*b, funcs...)
}

// Build returns a new Scheme containing the registered types.
func (b *Builder) Build() (*Scheme, error) {
	s := New()
	if err := b.AddToScheme(s); err != nil {
		return nil, err
	}
	return s, nil
}
