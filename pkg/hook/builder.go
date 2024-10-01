package hook

// Builder is a helper for constructing Hook instances with registered hooks.
type Builder []Register

var _ Register = (*Builder)(nil)

// NewBuilder creates a new Builder with optional initial hook functions.
func NewBuilder(registers ...Register) Builder {
	return Builder(registers)
}

// AddToHooks adds all registered hook functions to the provided Hook instance.
func (b Builder) AddToHooks(hook *Hook) error {
	for _, f := range b {
		if err := f.AddToHooks(hook); err != nil {
			return err
		}
	}
	return nil
}

// Register appends one or more hook functions to the Builder.
func (b *Builder) Register(registers ...Register) {
	*b = append(*b, registers...)
}

// Build creates a new Hook instance and adds all registered hook functions to it.
func (b Builder) Build() (*Hook, error) {
	h := New()
	if err := b.AddToHooks(h); err != nil {
		return nil, err
	}
	return h, nil
}
