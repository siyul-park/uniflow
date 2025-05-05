package hook

// Builder is a helper for constructing Hook instances with registered hooks.
type Builder []Register

var _ Register = (*Builder)(nil)

// NewBuilder creates a new Builder with optional initial hook functions.
func NewBuilder(registers ...Register) *Builder {
	b := &Builder{}
	for _, r := range registers {
		b.Register(r)
	}
	return b
}

// AddToHook adds all registered hook functions to the provided Hook instance.
func (b *Builder) AddToHook(hook *Hook) error {
	for _, f := range *b {
		if err := f.AddToHook(hook); err != nil {
			return err
		}
	}
	return nil
}

// Register appends one or more hook functions to the Builder.
func (b *Builder) Register(register Register) bool {
	for _, r := range *b {
		if r == register {
			return false
		}
	}
	*b = append(*b, register)
	return true
}

// Unregister removes a hook function from the Builder.
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

// Build creates a new Hook instance and adds all registered hook functions to it.
func (b *Builder) Build() (*Hook, error) {
	h := New()
	if err := b.AddToHook(h); err != nil {
		return nil, err
	}
	return h, nil
}
