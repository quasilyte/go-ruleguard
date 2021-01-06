// Package types mimics the https://golang.org/pkg/go/types/ package.
// It also contains some extra utility functions, they're defined in ext.go file.
package types

// Implements reports whether a given type implements the specified interface.
func Implements(typ Type, iface *Interface) bool { return false }

// Identical reports whether x and y are identical types. Receivers of Signature types are ignored.
func Identical(x, y Type) bool { return false }

// A Type represents a type of Go. All types implement the Type interface.
type Type interface {
	// Underlying returns the underlying type of a type.
	Underlying() Type

	// String returns a string representation of a type.
	String() string
}

type (
	// A Pointer represents a pointer type.
	Pointer struct{}

	// An Interface represents an interface type.
	Interface struct{}
)

// NewPointer returns a new pointer type for the given element (base) type.
func NewPointer(elem Type) *Pointer { return nil }

// Elem returns the element type for the given pointer.
func (*Pointer) Elem() Type { return nil }
