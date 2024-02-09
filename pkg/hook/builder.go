package hook

// Builder is a helper for building Hooks instances with registered hooks.
type Builder []func(*Hook) error

// NewBuilder creates a new Builder with optional initial hook functions.
func NewBuilder(funcs ...func(*Hook) error) Builder {
	return Builder(funcs)
}

// AddToHooks adds all registered hook functions to the provided Hook instance.
func (b Builder) AddToHooks(h *Hook) error {
	for _, f := range b {
		if err := f(h); err != nil {
			return err
		}
	}
	return nil
}

// Register registers one or more hook functions to the Builder.
func (b *Builder) Register(funcs ...func(*Hook) error) {
	*b = append(*b, funcs...)
}

// Build creates a new Hook instance containing all registered hooks.
func (b Builder) Build() (*Hook, error) {
	h := New()
	if err := b.AddToHooks(h); err != nil {
		return nil, err
	}
	return h, nil
}
