// +build ignore

package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
	"github.com/quasilyte/go-ruleguard/dsl/types"
)

func stringUnderlying(ctx *dsl.VarFilterContext) bool {
	// Test both Type.Underlying() and Type.String() methods.
	return ctx.Type.Underlying().String() == `string`
}

func isZeroSize(ctx *dsl.VarFilterContext) bool {
	return ctx.SizeOf(ctx.Type) == 0
}

func isPointer(ctx *dsl.VarFilterContext) bool {
	// There is no Type.IsT() methods (yet?), but it's possible to
	// use nil comparison for that.
	ptr := types.AsPointer(ctx.Type)
	return ptr != nil
}

func isInterface(ctx *dsl.VarFilterContext) bool {
	// Nil can be used on either side.
	return nil != types.AsInterface(ctx.Type.Underlying())
}

func isError(ctx *dsl.VarFilterContext) bool {
	// Testing Interface.String() method.
	iface := types.AsInterface(ctx.Type.Underlying())
	if iface != nil {
		return iface.String() == `interface{Error() string}`
	}
	return false
}

func isInterfacePtr(ctx *dsl.VarFilterContext) bool {
	ptr := types.AsPointer(ctx.Type)
	if ptr != nil {
		return types.AsInterface(ptr.Elem().Underlying()) != nil
	}
	return false
}

func typeNameHasErrorSuffix(ctx *dsl.VarFilterContext) bool {
	// Test string operations; this is basically strings.HasSuffix().
	s := ctx.Type.String()
	suffix := "Error"
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func implementsStringer(ctx *dsl.VarFilterContext) bool {
	// Using a non-constant argument to GetInterface() on purpose.
	ifaceName := `fmt.Stringer`
	stringer := ctx.GetInterface(ifaceName)
	return types.Implements(ctx.Type, stringer) ||
		types.Implements(types.NewPointer(ctx.Type), stringer)
}

func ptrElemSmallerThanUintptr(ctx *dsl.VarFilterContext) bool {
	ptr := types.AsPointer(ctx.Type)
	if ptr == nil {
		return false // Not a pointer
	}
	uintptrSize := ctx.SizeOf(ctx.GetType(`uintptr`))
	elemSize := ctx.SizeOf(ptr.Elem())
	return elemSize < uintptrSize
}

func testRules(m dsl.Matcher) {
	m.Match(`test($x, "underlying type is string")`).
		Where(m["x"].Filter(stringUnderlying)).
		Report(`true`)

	m.Match(`test($x, "zero sized")`).
		Where(m["x"].Filter(isZeroSize)).
		Report(`true`)

	m.Match(`test($x, "type is pointer")`).
		Where(m["x"].Filter(isPointer)).
		Report(`true`)

	m.Match(`test($x, "type is error")`).
		Where(m["x"].Filter(isError)).
		Report(`true`)

	// Use a custom filter negation.
	m.Match(`test($x, "type is not interface")`).
		Where(!m["x"].Filter(isInterface)).
		Report(`true`)

	m.Match(`test($x, "type name has Error suffix")`).
		Where(m["x"].Filter(typeNameHasErrorSuffix)).
		Report(`true`)

	m.Match(`test($x, "implements fmt.Stringer")`).
		Where(m["x"].Filter(implementsStringer)).
		Report(`true`)

	m.Match(`test($x, "pointer to interface")`).
		Where(m["x"].Filter(isInterfacePtr)).
		Report(`true`)

	m.Match(`test($x, "pointer elem value size is smaller than uintptr")`).
		Where(m["x"].Filter(ptrElemSmallerThanUintptr)).
		Report(`true`)
}
