package dsl

// MatchResult holds all recent match vars.
type MatchResult map[string]Var

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

// Match specifies a set of patterns that match current rule.
// Pattern matching succeeds if at least 1 pattern matches.
//
// If none of the given patterns matched, rule execution terminates.
func Match(pattern string, alternatives ...string) {}

// Filter applies additional constraint that make a recent
// match to be either accepted or rejected.
func Filter(cond bool) {}

// Error yields an error-level message if a recent match was accepted.
func Error(message ruleMessage) {}

// Hint yields an hint-level message if a recent match was accepted.
func Hint(message ruleMessage) {}

// ruleMessage is a string that can contain interpolated expressions.
//
// For every matched variable it's possible to interpolate
// their printed representation into the message text with $<name>.
// An entire match can be addressed with $$.
type ruleMessage string

type exprType struct{}

// AssignableTo reports whether a type is assign-compatible with a given type.
func (exprType) AssignableTo(typ string) bool { return boolResult }

// Is reports whether a type is identical to a given type.
func (exprType) Is(typ string) bool { return boolResult }
