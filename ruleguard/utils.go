package ruleguard

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"strings"
)

func sprintNode(fset *token.FileSet, n ast.Node) string {
	if fset == nil {
		fset = token.NewFileSet()
	}
	var buf strings.Builder
	if err := printer.Fprint(&buf, fset, n); err != nil {
		return ""
	}
	return buf.String()
}

func typeFromNode(e ast.Expr) types.Type {
	switch e := e.(type) {
	case *ast.Ident:
		switch e.Name {
		case "string":
			return types.Typ[types.String]
		case "int":
			return types.Typ[types.Int]
		}
	}
	panic(fmt.Sprintf("can't convert %T to a type", e))
}

// isPure reports whether expr is a softly safe expression and contains
// no significant side-effects. As opposed to strictly safe expressions,
// soft safe expressions permit some forms of side-effects, like
// panic possibility during indexing or nil pointer dereference.
//
// Uses types info to determine type conversion expressions that
// are the only permitted kinds of call expressions.
// Note that is does not check whether called function really
// has any side effects. The analysis is very conservative.
func isPure(info *types.Info, expr ast.Expr) bool {
	// This list switch is not comprehensive and uses
	// whitelist to be on the conservative side.
	// Can be extended as needed.

	switch expr := expr.(type) {
	case *ast.StarExpr:
		return isPure(info, expr.X)
	case *ast.BinaryExpr:
		return isPure(info, expr.X) &&
			isPure(info, expr.Y)
	case *ast.UnaryExpr:
		return expr.Op != token.ARROW &&
			isPure(info, expr.X)
	case *ast.BasicLit, *ast.Ident:
		return true
	case *ast.IndexExpr:
		return isPure(info, expr.X) &&
			isPure(info, expr.Index)
	case *ast.SelectorExpr:
		return isPure(info, expr.X)
	case *ast.ParenExpr:
		return isPure(info, expr.X)
	case *ast.CompositeLit:
		return isPureList(info, expr.Elts)
	case *ast.CallExpr:
		return isTypeExpr(info, expr.Fun) && isPureList(info, expr.Args)

	default:
		return false
	}
}

// isPureList reports whether every expr in list is safe.
//
// See isPure.
func isPureList(info *types.Info, list []ast.Expr) bool {
	for _, expr := range list {
		if !isPure(info, expr) {
			return false
		}
	}
	return true
}

func isConstant(info *types.Info, expr ast.Expr) bool {
	tv, ok := info.Types[expr]
	return ok && tv.Value != nil
}

// isTypeExpr reports whether x represents a type expression.
//
// Type expression does not evaluate to any run time value,
// but rather describes a type that is used inside Go expression.
//
// For example, (*T)(v) is a CallExpr that "calls" (*T).
// (*T) is a type expression that tells Go compiler type v should be converted to.
func isTypeExpr(info *types.Info, x ast.Expr) bool {
	switch x := x.(type) {
	case *ast.StarExpr:
		return isTypeExpr(info, x.X)
	case *ast.ParenExpr:
		return isTypeExpr(info, x.X)
	case *ast.SelectorExpr:
		return isTypeExpr(info, x.Sel)

	case *ast.Ident:
		// Identifier may be a type expression if object
		// it reffers to is a type name.
		_, ok := info.ObjectOf(x).(*types.TypeName)
		return ok

	case *ast.FuncType, *ast.StructType, *ast.InterfaceType, *ast.ArrayType, *ast.MapType, *ast.ChanType:
		return true

	default:
		return false
	}
}
