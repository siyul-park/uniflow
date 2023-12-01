package hook

// Builder builds a new Hooks.
type Builder []func(*Hook) error

// NewBuilder returns a new HooksBuilder.
func NewBuilder(funcs ...func(*Hook) error) Builder {
	return Builder(funcs)
}

// AddToHooks adds all registered hooks to h.
func (b Builder) AddToHooks(h *Hook) error {
	for _, f := range b {
		if err := f(h); err != nil {
			return err
		}
	}
	return nil
}

// Register adds one or more hooks.
func (b *Builder) Register(funcs ...func(*Hook) error) {
	*b = append(*b, funcs...)
}

// Build returns a new Hooks containing the registered hooks.
func (b Builder) Build() (*Hook, error) {
	h := New()
	if err := b.AddToHooks(h); err != nil {
		return nil, err
	}
	return h, nil
}
