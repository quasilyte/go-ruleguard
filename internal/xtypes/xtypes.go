package xtypes

import (
	"go/types"

	"golang.org/x/exp/typeparams"
)

// Implements reports whether type v implements iface.
//
// Unlike types.Implements(), it permits X and Y named types
// to be considered identical even if their addresses are different,
// and supports matching generic interfaces with type parameters.
func Implements(v types.Type, iface *types.Interface) bool {
	if iface.Empty() {
		return true
	}

	if interfaceHasTypeParams(iface) {
		return implementsGeneric(v, iface)
	}

	if types.Implements(v, iface) {
		return true
	}

	return implementsMethods(v, iface)
}

func implementsMethods(v types.Type, iface *types.Interface) bool {
	if v, _ := v.Underlying().(*types.Interface); v != nil {
		for i := 0; i < iface.NumMethods(); i++ {
			m := iface.Method(i)
			obj, _, _ := types.LookupFieldOrMethod(v, false, m.Pkg(), m.Name())
			switch {
			case obj == nil:
				return false
			case !Identical(obj.Type(), m.Type()):
				return false
			}
		}
		return checkEmbeddeds(v, iface)
	}

	// A concrete type v implements iface if it implements all methods of iface.
	for i := 0; i < iface.NumMethods(); i++ {
		m := iface.Method(i)

		obj, _, _ := types.LookupFieldOrMethod(v, false, m.Pkg(), m.Name())
		if obj == nil {
			return false
		}

		f, ok := obj.(*types.Func)
		if !ok {
			return false
		}

		if !Identical(f.Type(), m.Type()) {
			return false
		}
	}

	return checkEmbeddeds(v, iface)
}

func checkEmbeddeds(v types.Type, iface *types.Interface) bool {
	for i := 0; i < iface.NumEmbeddeds(); i++ {
		emb := iface.EmbeddedType(i)
		if emb.String() == "comparable" {
			if !types.Comparable(v) {
				return false
			}
			continue
		}
		if embIface, ok := emb.Underlying().(*types.Interface); ok {
			if !Implements(v, embIface) {
				return false
			}
		}
	}
	return true
}

func implementsGeneric(v types.Type, iface *types.Interface) bool {
	subst := make(map[*typeparams.TypeParam]types.Type)

	for i := 0; i < iface.NumMethods(); i++ {
		m := iface.Method(i)
		obj, _, _ := types.LookupFieldOrMethod(v, false, m.Pkg(), m.Name())
		if obj == nil {
			return false
		}
		f, ok := obj.(*types.Func)
		if !ok {
			return false
		}
		if !unifyTypes(m.Type(), f.Type(), subst) {
			return false
		}
	}

	methods := make([]*types.Func, iface.NumMethods())
	for i := 0; i < iface.NumMethods(); i++ {
		m := iface.Method(i)
		newSig := substType(m.Type(), subst)
		sig, ok := newSig.(*types.Signature)
		if !ok {
			return false
		}
		methods[i] = types.NewFunc(m.Pos(), m.Pkg(), m.Name(), sig)
	}

	var embeddeds []types.Type
	for i := 0; i < iface.NumEmbeddeds(); i++ {
		embeddeds = append(embeddeds, substType(iface.EmbeddedType(i), subst))
	}

	unrolled := types.NewInterfaceType(methods, embeddeds)
	unrolled.Complete()

	if types.Implements(v, unrolled) {
		return true
	}

	return implementsMethods(v, unrolled)
}

func interfaceHasTypeParams(iface *types.Interface) bool {
	for i := 0; i < iface.NumMethods(); i++ {
		if hasTypeParams(iface.Method(i).Type()) {
			return true
		}
	}
	return false
}

func hasTypeParams(t types.Type) bool {
	switch t := t.(type) {
	case *typeparams.TypeParam:
		return true
	case *types.Named:
		targs := t.TypeArgs()
		if targs != nil {
			for i := 0; i < targs.Len(); i++ {
				if hasTypeParams(targs.At(i)) {
					return true
				}
			}
		}
		return false
	case *types.Pointer:
		return hasTypeParams(t.Elem())
	case *types.Slice:
		return hasTypeParams(t.Elem())
	case *types.Array:
		return hasTypeParams(t.Elem())
	case *types.Map:
		return hasTypeParams(t.Key()) || hasTypeParams(t.Elem())
	case *types.Chan:
		return hasTypeParams(t.Elem())
	case *types.Signature:
		if t.Params() != nil {
			for i := 0; i < t.Params().Len(); i++ {
				if hasTypeParams(t.Params().At(i).Type()) {
					return true
				}
			}
		}
		if t.Results() != nil {
			for i := 0; i < t.Results().Len(); i++ {
				if hasTypeParams(t.Results().At(i).Type()) {
					return true
				}
			}
		}
		return false
	case *types.Tuple:
		for i := 0; i < t.Len(); i++ {
			if hasTypeParams(t.At(i).Type()) {
				return true
			}
		}
		return false
	case *types.Interface:
		for i := 0; i < t.NumMethods(); i++ {
			if hasTypeParams(t.Method(i).Type()) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func unifyTypes(pattern, concrete types.Type, subst map[*typeparams.TypeParam]types.Type) bool {
	if tp, ok := pattern.(*typeparams.TypeParam); ok {
		if existing, ok := subst[tp]; ok {
			return Identical(existing, concrete)
		}
		subst[tp] = concrete
		return true
	}

	if p, ok := pattern.(*types.Pointer); ok {
		c, ok := concrete.(*types.Pointer)
		if !ok {
			return false
		}
		return unifyTypes(p.Elem(), c.Elem(), subst)
	}

	if p, ok := pattern.(*types.Slice); ok {
		c, ok := concrete.(*types.Slice)
		if !ok {
			return false
		}
		return unifyTypes(p.Elem(), c.Elem(), subst)
	}

	if p, ok := pattern.(*types.Array); ok {
		c, ok := concrete.(*types.Array)
		if !ok || p.Len() != c.Len() {
			return false
		}
		return unifyTypes(p.Elem(), c.Elem(), subst)
	}

	if p, ok := pattern.(*types.Map); ok {
		c, ok := concrete.(*types.Map)
		if !ok {
			return false
		}
		return unifyTypes(p.Key(), c.Key(), subst) && unifyTypes(p.Elem(), c.Elem(), subst)
	}

	if p, ok := pattern.(*types.Chan); ok {
		c, ok := concrete.(*types.Chan)
		if !ok || p.Dir() != c.Dir() {
			return false
		}
		return unifyTypes(p.Elem(), c.Elem(), subst)
	}

	if p, ok := pattern.(*types.Signature); ok {
		c, ok := concrete.(*types.Signature)
		if !ok || p.Variadic() != c.Variadic() {
			return false
		}
		pParams := p.Params()
		cParams := c.Params()
		pLen := 0
		if pParams != nil {
			pLen = pParams.Len()
		}
		cLen := 0
		if cParams != nil {
			cLen = cParams.Len()
		}
		if pLen != cLen {
			return false
		}
		for i := 0; i < pLen; i++ {
			if !unifyTypes(pParams.At(i).Type(), cParams.At(i).Type(), subst) {
				return false
			}
		}
		pRes := p.Results()
		cRes := c.Results()
		pRLen := 0
		if pRes != nil {
			pRLen = pRes.Len()
		}
		cRLen := 0
		if cRes != nil {
			cRLen = cRes.Len()
		}
		if pRLen != cRLen {
			return false
		}
		for i := 0; i < pRLen; i++ {
			if !unifyTypes(pRes.At(i).Type(), cRes.At(i).Type(), subst) {
				return false
			}
		}
		return true
	}

	if p, ok := pattern.(*types.Tuple); ok {
		c, ok := concrete.(*types.Tuple)
		if !ok || p.Len() != c.Len() {
			return false
		}
		for i := 0; i < p.Len(); i++ {
			if !unifyTypes(p.At(i).Type(), c.At(i).Type(), subst) {
				return false
			}
		}
		return true
	}

	if p, ok := pattern.(*types.Named); ok {
		c, ok := concrete.(*types.Named)
		if !ok {
			return false
		}
		if p.Obj().Name() != c.Obj().Name() {
			return false
		}
		pTargs := p.TypeArgs()
		cTargs := c.TypeArgs()
		if pTargs != nil && cTargs != nil {
			if pTargs.Len() != cTargs.Len() {
				return false
			}
			for i := 0; i < pTargs.Len(); i++ {
				if !unifyTypes(pTargs.At(i), cTargs.At(i), subst) {
					return false
				}
			}
			return true
		}
		return Identical(p, c)
	}

	return Identical(pattern, concrete)
}

func substType(t types.Type, subst map[*typeparams.TypeParam]types.Type) types.Type {
	if tp, ok := t.(*typeparams.TypeParam); ok {
		if repl, ok := subst[tp]; ok {
			return repl
		}
		if tp.Constraint() != nil && tp.Constraint().String() == "comparable" {
			return types.Typ[types.Int]
		}
		return types.NewInterfaceType(nil, nil)
	}

	if ptr, ok := t.(*types.Pointer); ok {
		return types.NewPointer(substType(ptr.Elem(), subst))
	}

	if sl, ok := t.(*types.Slice); ok {
		return types.NewSlice(substType(sl.Elem(), subst))
	}

	if arr, ok := t.(*types.Array); ok {
		return types.NewArray(substType(arr.Elem(), subst), arr.Len())
	}

	if m, ok := t.(*types.Map); ok {
		return types.NewMap(substType(m.Key(), subst), substType(m.Elem(), subst))
	}

	if ch, ok := t.(*types.Chan); ok {
		return types.NewChan(ch.Dir(), substType(ch.Elem(), subst))
	}

	if sig, ok := t.(*types.Signature); ok {
		var params []*types.Var
		if sig.Params() != nil {
			for i := 0; i < sig.Params().Len(); i++ {
				v := sig.Params().At(i)
				params = append(params, types.NewVar(v.Pos(), v.Pkg(), v.Name(), substType(v.Type(), subst)))
			}
		}
		var results []*types.Var
		if sig.Results() != nil {
			for i := 0; i < sig.Results().Len(); i++ {
				v := sig.Results().At(i)
				results = append(results, types.NewVar(v.Pos(), v.Pkg(), v.Name(), substType(v.Type(), subst)))
			}
		}
		return types.NewSignatureType(sig.Recv(), nil, nil, types.NewTuple(params...), types.NewTuple(results...), sig.Variadic())
	}

	if named, ok := t.(*types.Named); ok {
		targs := named.TypeArgs()
		if targs != nil && targs.Len() > 0 {
			var newArgs []types.Type
			changed := false
			for i := 0; i < targs.Len(); i++ {
				origArg := targs.At(i)
				newArg := substType(origArg, subst)
				if newArg != origArg {
					changed = true
				}
				newArgs = append(newArgs, newArg)
			}
			if changed {
				inst, err := types.Instantiate(nil, named.Origin(), newArgs, false)
				if err == nil {
					return inst
				}
			}
		}
		return named
	}

	if tup, ok := t.(*types.Tuple); ok {
		var vars []*types.Var
		for i := 0; i < tup.Len(); i++ {
			v := tup.At(i)
			vars = append(vars, types.NewVar(v.Pos(), v.Pkg(), v.Name(), substType(v.Type(), subst)))
		}
		return types.NewTuple(vars...)
	}

	return t
}


// Identical reports whether x and y are identical types.
//
// Unlike types.Identical(), it permits X and Y named types
// to be considered identical even if their addresses are different.
func Identical(x, y types.Type) bool {
	return typeIdentical(x, y, nil)
}

func typeIdentical(x, y types.Type, p *ifacePair) bool {
	if x == y {
		return true
	}

	switch x := x.(type) {
	case *types.Basic:
		// Basic types are singletons except for the rune and byte
		// aliases, thus we cannot solely rely on the x == y check
		// above. See also comment in TypeName.IsAlias.
		if y, ok := y.(*types.Basic); ok {
			return x.Kind() == y.Kind()
		}

	case *types.Array:
		// Two array types are identical if they have identical element types
		// and the same array length.
		if y, ok := y.(*types.Array); ok {
			// If one or both array lengths are unknown (< 0) due to some error,
			// assume they are the same to avoid spurious follow-on errors.
			return (x.Len() < 0 || y.Len() < 0 || x.Len() == y.Len()) && typeIdentical(x.Elem(), y.Elem(), p)
		}

	case *types.Slice:
		// Two slice types are identical if they have identical element types.
		if y, ok := y.(*types.Slice); ok {
			return typeIdentical(x.Elem(), y.Elem(), p)
		}

	case *types.Struct:
		// Two struct types are identical if they have the same sequence of fields,
		// and if corresponding fields have the same names, and identical types,
		// and identical tags. Two embedded fields are considered to have the same
		// name. Lower-case field names from different packages are always different.
		if y, ok := y.(*types.Struct); ok {
			if x.NumFields() == y.NumFields() {
				for i := 0; i < x.NumFields(); i++ {
					f := x.Field(i)
					g := y.Field(i)
					if f.Embedded() != g.Embedded() || !sameID(f, g.Pkg(), g.Name()) || !typeIdentical(f.Type(), g.Type(), p) {
						return false
					}
				}
				return true
			}
		}

	case *types.Pointer:
		// Two pointer types are identical if they have identical base types.
		if y, ok := y.(*types.Pointer); ok {
			return typeIdentical(x.Elem(), y.Elem(), p)
		}

	case *types.Tuple:
		// Two tuples types are identical if they have the same number of elements
		// and corresponding elements have identical types.
		if y, ok := y.(*types.Tuple); ok {
			if x.Len() == y.Len() {
				if x != nil {
					for i := 0; i < x.Len(); i++ {
						v := x.At(i)
						w := y.At(i)
						if !typeIdentical(v.Type(), w.Type(), p) {
							return false
						}
					}
				}
				return true
			}
		}

	case *types.Signature:
		// Two function types are identical if they have the same number of parameters
		// and result values, corresponding parameter and result types are identical,
		// and either both functions are variadic or neither is. Parameter and result
		// names are not required to match.
		if y, ok := y.(*types.Signature); ok {
			return x.Variadic() == y.Variadic() &&
				typeIdentical(x.Params(), y.Params(), p) &&
				typeIdentical(x.Results(), y.Results(), p)
		}

	case *typeparams.Union:
		// TODO(quasilyte): do we want to match generic union types too?
		// It would require copying a lot of code from the go/types.
		return false

	case *types.Interface:
		// Two interface types are identical if they have the same set of methods with
		// the same names and identical function types. Lower-case method names from
		// different packages are always different. The order of the methods is irrelevant.
		if y, ok := y.(*types.Interface); ok {
			if x.NumMethods() != y.NumMethods() {
				return false
			}
			// Interface types are the only types where cycles can occur
			// that are not "terminated" via named types; and such cycles
			// can only be created via method parameter types that are
			// anonymous interfaces (directly or indirectly) embedding
			// the current interface. Example:
			//
			//    type T interface {
			//        m() interface{T}
			//    }
			//
			// If two such (differently named) interfaces are compared,
			// endless recursion occurs if the cycle is not detected.
			//
			// If x and y were compared before, they must be equal
			// (if they were not, the recursion would have stopped);
			// search the ifacePair stack for the same pair.
			//
			// This is a quadratic algorithm, but in practice these stacks
			// are extremely short (bounded by the nesting depth of interface
			// type declarations that recur via parameter types, an extremely
			// rare occurrence). An alternative implementation might use a
			// "visited" map, but that is probably less efficient overall.
			q := &ifacePair{x, y, p}
			for p != nil {
				if p.identical(q) {
					return true // same pair was compared before
				}
				p = p.prev
			}
			for i := 0; i < x.NumMethods(); i++ {
				f := x.Method(i)
				g := y.Method(i)
				if f.Id() != g.Id() || !typeIdentical(f.Type(), g.Type(), q) {
					return false
				}
			}
			return true
		}

	case *types.Map:
		// Two map types are identical if they have identical key and value types.
		if y, ok := y.(*types.Map); ok {
			return typeIdentical(x.Key(), y.Key(), p) && typeIdentical(x.Elem(), y.Elem(), p)
		}

	case *types.Chan:
		// Two channel types are identical if they have identical value types
		// and the same direction.
		if y, ok := y.(*types.Chan); ok {
			return x.Dir() == y.Dir() && typeIdentical(x.Elem(), y.Elem(), p)
		}

	case *types.Named:
		// Two named types are identical if their type names originate
		// in the same type declaration.
		y, ok := y.(*types.Named)
		if !ok {
			return false
		}
		if x.Obj() == y.Obj() {
			return true
		}
		return sameID(x.Obj(), y.Obj().Pkg(), y.Obj().Name())

	case *typeparams.TypeParam:
		// nothing to do (x and y being equal is caught in the very beginning of this function)

	case *types.Alias:
		// an alias type is identical if the type it's an alias of is identical to it.
		return typeIdentical(types.Unalias(x), y, p)

	case nil:
		// avoid a crash in case of nil type

	default:
		panic("unreachable")
	}

	return false
}

// An ifacePair is a node in a stack of interface type pairs compared for identity.
type ifacePair struct {
	x    *types.Interface
	y    *types.Interface
	prev *ifacePair
}

func (p *ifacePair) identical(q *ifacePair) bool {
	return (p.x == q.x && p.y == q.y) ||
		(p.x == q.y && p.y == q.x)
}

func sameID(obj types.Object, pkg *types.Package, name string) bool {
	// spec:
	// "Two identifiers are different if they are spelled differently,
	// or if they appear in different packages and are not exported.
	// Otherwise, they are the same."
	if name != obj.Name() {
		return false
	}
	// obj.Name == name
	if obj.Exported() {
		return true
	}
	// not exported, so packages must be the same (pkg == nil for
	// fields in Universe scope; this can only happen for types
	// introduced via Eval)
	if pkg == nil || obj.Pkg() == nil {
		return pkg == obj.Pkg()
	}
	// pkg != nil && obj.pkg != nil
	return pkg.Path() == obj.Pkg().Path()
}
