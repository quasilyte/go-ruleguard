//go:build ignore
// +build ignore

package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
	"github.com/quasilyte/go-ruleguard/dsl/types"
)

func embedsMutex(ctx *dsl.VarFilterContext) bool {
	typ := ctx.Type.Underlying()
	asPointer := types.AsPointer(typ)
	if asPointer != nil {
		typ = asPointer.Elem().Underlying()
	}
	asStruct := types.AsStruct(typ)
	if asStruct == nil {
		return false
	}
	mutexType := ctx.GetType(`sync.Mutex`)
	i := 0
	for i < asStruct.NumFields() {
		field := asStruct.Field(i)
		if field.Embedded() && types.Identical(field.Type(), mutexType) {
			return true
		}
		i++
	}
	return false
}

func derefPointer(ptr *types.Pointer) *types.Pointer {
	return types.AsPointer(ptr.Elem())
}

func tooManyPointers(ctx *dsl.VarFilterContext) bool {
	indir := 0
	ptr := types.AsPointer(ctx.Type)
	for ptr != nil {
		indir++
		ptr = derefPointer(ptr)
	}
	return indir >= 3
}

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

func isInterfaceImpl(ctx *dsl.VarFilterContext) bool {
	// Nil can be used on either side.
	return nil != types.AsInterface(ctx.Type.Underlying())
}

func isInterface(ctx *dsl.VarFilterContext) bool {
	// Forwarding a call to other function.
	return isInterfaceImpl(ctx)
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

func isIntArray3(ctx *dsl.VarFilterContext) bool {
	arr3 := types.NewArray(ctx.GetType(`int`), 3)
	return types.Identical(ctx.Type, arr3)
}

func isIntArray(ctx *dsl.VarFilterContext) bool {
	arr := types.AsArray(ctx.Type)
	if arr != nil {
		return types.Identical(ctx.GetType(`int`), arr.Elem())
	}
	return false
}

func isIntSlice(ctx *dsl.VarFilterContext) bool {
	intSlice := types.NewSlice(ctx.GetType(`int`))
	return types.Identical(ctx.Type, intSlice)
}

func testRules(m dsl.Matcher) {
	m.Match(`test($x, "is [3]int")`).
		Where(m["x"].Filter(isIntArray3)).
		Report(`true`)

	m.Match(`test($x, "is int array")`).
		Where(m["x"].Filter(isIntArray)).
		Report(`true`)

	m.Match(`test($x, "is int slice")`).
		Where(m["x"].Filter(isIntSlice)).
		Report(`true`)

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

	m.Match(`test($x, "indirection of 3 or more pointers")`).
		Where(m["x"].Filter(tooManyPointers)).
		Report(`true`)

	m.Match(`test($x, "embeds a mutex")`).
		Where(m["x"].Filter(embedsMutex)).
		Report(`true`)
}
