package fluent

// Matcher is a main API group-level entry point.
// It's used to define and configure the group rules.
// It also represents a map of all rule-local variables.
type Matcher map[string]Var

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
func (Matcher) Report(message string) {}

// Var is a pattern variable that describes a named submatch.
type Var struct {
	// Pure reports whether expr matched by var is side-effect-free.
	Pure bool

	// Const reports whether expr matched by var is a constant value.
	Const bool

	// Type is a type of a matched expr.
	Type exprType
}

// IsPure asserts that expression matched by v is side-effect-free.
func (v Var) IsPure() {}

type exprType struct{}

// AssignableTo reports whether a type is assign-compatible with a given type.
func (exprType) AssignableTo(typ string) bool { return boolResult }

// Is reports whether a type is identical to a given type.
func (exprType) Is(typ string) bool { return boolResult }
