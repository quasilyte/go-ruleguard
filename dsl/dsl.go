package dsl

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

// File returns the current file context.
func (m Matcher) File() File { return File{} }

// Var is a pattern variable that describes a named submatch.
type Var struct {
	// Pure reports whether expr matched by var is side-effect-free.
	Pure bool

	// Const reports whether expr matched by var is a constant value.
	Const bool

	// Value is a compile-time computable value of the expression.
	Value ExprValue

	// Addressable reports whether the corresponding expression is addressable.
	// See https://golang.org/ref/spec#Address_operators.
	Addressable bool

	// Type is a type of a matched expr.
	//
	// For function call expressions, a type is a function result type,
	// but for a function expression itself it's a *types.Signature.
	//
	// Suppose we have a `a.b()` expression:
	//	`$x()` m["x"].Type is `a.b` function type
	//	`$x` m["x"].Type is `a.b()` function call result type
	Type ExprType

	// Text is a captured node text as in the source code.
	Text MatchedText

	// Node is a captured AST node.
	Node MatchedNode
}

// MatchedNode represents an AST node associated with a named submatch.
type MatchedNode struct{}

// Is reports whether a matched node AST type is compatible with the specified type.
// A valid argument is a ast.Node implementing type name from the "go/ast" package.
// Examples: "BasicLit", "Expr", "Stmt", "Ident", "ParenExpr".
// See https://golang.org/pkg/go/ast/.
func (MatchedNode) Is(typ string) bool { return boolResult }

// ExprValue describes a compile-time computable value of a matched expr.
type ExprValue struct{}

// Int returns compile-time computable int value of the expression.
// If value can't be computed, condition will fail.
func (ExprValue) Int() int { return intResult }

// ExprType describes a type of a matcher expr.
type ExprType struct {
	// Size represents expression type size in bytes.
	Size int
}

// Underlying returns expression type underlying type.
// See https://golang.org/pkg/go/types/#Type Underlying() method documentation.
// Read https://golang.org/ref/spec#Types section to learn more about underlying types.
func (ExprType) Underlying() ExprType { return underlyingType }

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

// MatchedText represents a source text associated with a matched node.
type MatchedText string

// Matches reports whether the text matches the given regexp pattern.
func (MatchedText) Matches(pattern string) bool { return boolResult }

// String represents an arbitrary string-typed data.
type String string

// Matches reports whether a string matches the given regexp pattern.
func (String) Matches(pattern string) bool { return boolResult }

// File represents the current Go source file.
type File struct {
	// Name is a file base name.
	Name String

	// PkgPath is a file package path.
	// Examples: "io/ioutil", "strings", "github.com/quasilyte/go-ruleguard/dsl".
	PkgPath String
}

// Imports reports whether the current file imports the given path.
func (File) Imports(path string) bool { return boolResult }
