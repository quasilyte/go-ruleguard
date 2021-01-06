package types

// AsPointer is a type-assert like operation, x.(*Pointer), but never panics.
// Returns nil if type is not a pointer.
func AsPointer(x Type) *Pointer { return nil }

// AsInterface is a type-assert like operation, x.(*Interface), but never panics.
// Returns nil if type is not an interface.
func AsInterface(x Type) *Interface { return nil }
