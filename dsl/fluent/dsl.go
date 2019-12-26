package fluent

// Matcher is a main API group-level entry point.
// It's used to define and configure the group rules.
// It also represents a map of all rule-local variables.
type Matcher map[string]Var

// Import loads given package path into a rule group imports table.
//
// That table is used during the rules compilation.
//
// The table has the following effect on the rules:
//	* For type expressions, it's used to resolve the
//	  full package paths of qualified types, like `foo.Bar`.
//	  If Import(`a/b/foo`) is called, `foo.Bar` will match
//	  `a/b/foo.Bar` type during the pattern execution.
func (m Matcher) Import(pkgPath string) {}

// Match specifies a set of patterns that match a rule being defined.
// Pattern matching succeeds if at least 1 pattern matches.
//
// If none of the given patterns matched, rule execution stops.
func (m Matcher) Match(pattern string, alternatives ...string) Matcher {
	return m
}

// Where applies additional constraint to a match.
// If a given cond is not satisfied, a match is rejected and
// rule execution stops.
func (m Matcher) Where(cond bool) Matcher {
	return m
}

// Report prints a message if associated rule match is successful.
//
// A message is a string that can contain interpolated expressions.
// For every matched variable it's possible to interpolate
// their printed representation into the message text with $<name>.
// An entire match can be addressed with $$.
func (m Matcher) Report(message string) Matcher {
	return m
}

// Suggest assigns a quickfix suggestion for the matched code.
func (m Matcher) Suggest(suggestion string) Matcher {
	return m
}

// At binds the reported node to a named submatch.
// If no explicit location is given, the outermost node ($$) is used.
func (m Matcher) At(v Var) Matcher {
	return m
}

// Var is a pattern variable that describes a named submatch.
type Var struct {
	// Pure reports whether expr matched by var is side-effect-free.
	Pure bool

	// Const reports whether expr matched by var is a constant value.
	Const bool

	// Addressable reports whether the corresponding expression is addressable.
	// See https://golang.org/ref/spec#Address_operators.
	Addressable bool

	// Type is a type of a matched expr.
	Type ExprType
}

// ExprType describes a type of a matcher expr.
type ExprType struct {
	// Size represents expression type size in bytes.
	Size int
}

// AssignableTo reports whether a type is assign-compatible with a given type.
// See https://golang.org/pkg/go/types/#AssignableTo.
func (ExprType) AssignableTo(typ string) bool { return boolResult }

// ConvertibleTo reports whether a type is conversible to a given type.
// See https://golang.org/pkg/go/types/#ConvertibleTo.
func (ExprType) ConvertibleTo(typ string) bool { return boolResult }

// Implements reports whether a type implements a given interface.
// See https://golang.org/pkg/go/types/#Implements.
func (ExprType) Implements(typ string) bool { return boolResult }

// Is reports whether a type is identical to a given type.
func (ExprType) Is(typ string) bool { return boolResult }
