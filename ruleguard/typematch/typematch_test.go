package typematch

import (
	"go/token"
	"go/types"
	"path"
	"testing"
)

var (
	typeInt       = types.Typ[types.Int]
	typeString    = types.Typ[types.String]
	typeInt32     = types.Typ[types.Int32]
	typeUint8     = types.Typ[types.Uint8]
	typeUnsafePtr = types.Typ[types.UnsafePointer]
	typeEstruct   = types.NewStruct(nil, nil)

	stringerIface = types.NewInterfaceType([]*types.Func{
		types.NewFunc(token.NoPos, nil, "String",
			types.NewSignatureType(nil, nil, nil, types.NewTuple(), types.NewTuple(types.NewVar(token.NoPos, nil, "result", typeString)), false)),
	}, nil)

	intVar     = types.NewVar(token.NoPos, nil, "_", typeInt)
	int32Var   = types.NewVar(token.NoPos, nil, "_", typeInt32)
	estructVar = types.NewVar(token.NoPos, nil, "_", typeEstruct)
	stringVar  = types.NewVar(token.NoPos, nil, "_", typeString)

	testContext = &Context{
		Itab: NewImportsTab(map[string]string{
			"io":     "io",
			"syntax": "regexp/syntax",
		}),
	}
)

func structType(fields ...*types.Var) *types.Struct {
	return types.NewStruct(fields, nil)
}

func namedType2(pkgPath, typeName string) *types.Named {
	return namedType(pkgPath, path.Base(pkgPath), typeName)
}

func namedType(pkgPath, pkgName, typeName string) *types.Named {
	dummy := typeEstruct
	pkg := types.NewPackage(pkgPath, pkgName)
	typename := types.NewTypeName(0, pkg, typeName, dummy)
	return types.NewNamed(typename, dummy, nil)
}

func TestIdentical(t *testing.T) {
	tests := []struct {
		expr string
		typ  types.Type
	}{
		{`int`, typeInt},
		{`(int)`, typeInt},
		{`*int`, types.NewPointer(typeInt)},
		{`**int`, types.NewPointer(types.NewPointer(typeInt))},
		{`[]int`, types.NewSlice(typeInt)},
		{`[][]int`, types.NewSlice(types.NewSlice(typeInt))},
		{`[10]int`, types.NewArray(typeInt, 10)},
		{`map[int]int`, types.NewMap(typeInt, typeInt)},

		{`interface{}`, types.NewInterfaceType(nil, nil)},
		{`interface{ $*_ }`, stringerIface},

		{`$t`, typeInt},
		{`*$t`, types.NewPointer(typeInt)},
		{`*$t`, types.NewPointer(typeString)},
		{`*$t`, types.NewPointer(types.NewPointer(typeInt))},
		{`**$t`, types.NewPointer(types.NewPointer(typeInt))},
		{`map[$t]$t`, types.NewMap(typeInt, typeInt)},
		{`[$len]int`, types.NewArray(typeInt, 15)},
		{`[$len]int`, types.NewArray(typeInt, 20)},
		{`[$_]$_`, types.NewArray(typeInt, 20)},
		{`[$len][$len]int`, types.NewArray(types.NewArray(typeInt, 15), 15)},
		{`[$_][$_]int`, types.NewArray(types.NewArray(typeInt, 15), 10)},

		{`chan int`, types.NewChan(types.SendRecv, typeInt)},
		{`chan <- int`, types.NewChan(types.SendOnly, typeInt)},
		{`<- chan int`, types.NewChan(types.RecvOnly, typeInt)},
		{`chan $t`, types.NewChan(types.SendRecv, typeInt)},
		{`chan $t`, types.NewChan(types.SendRecv, typeString)},

		{`io.Reader`, namedType2("io", "Reader")},
		{`syntax.Regexp`, namedType2("regexp/syntax", "Regexp")},
		{`*syntax.Regexp`, types.NewPointer(namedType2("regexp/syntax", "Regexp"))},

		{`byte`, typeUint8},
		{`rune`, typeInt32},
		{`[]rune`, types.NewSlice(typeInt32)},
		{`[8]byte`, types.NewArray(typeUint8, 8)},

		{`unsafe.Pointer`, typeUnsafePtr},
		{`[]unsafe.Pointer`, types.NewSlice(typeUnsafePtr)},

		{`func()`, types.NewSignatureType(nil, nil, nil, nil, nil, false)},
		{`func(int)`, types.NewSignatureType(nil, nil, nil, types.NewTuple(intVar), nil, false)},
		{`func(int, string)`, types.NewSignatureType(nil, nil, nil, types.NewTuple(intVar, stringVar), nil, false)},
		{`func() int`, types.NewSignatureType(nil, nil, nil, nil, types.NewTuple(intVar), false)},
		{`func(string) int`, types.NewSignatureType(nil, nil, nil, types.NewTuple(stringVar), types.NewTuple(intVar), false)},
		{`func(int) int`, types.NewSignatureType(nil, nil, nil, types.NewTuple(intVar), types.NewTuple(intVar), false)},
		{`func() (string, int)`, types.NewSignatureType(nil, nil, nil, nil, types.NewTuple(stringVar, intVar), false)},

		{`func($_)`, types.NewSignatureType(nil, nil, nil, types.NewTuple(intVar), nil, false)},
		{`func($_)`, types.NewSignatureType(nil, nil, nil, types.NewTuple(stringVar), nil, false)},
		{`func($_) int`, types.NewSignatureType(nil, nil, nil, types.NewTuple(intVar), types.NewTuple(intVar), false)},
		{`func($_) int`, types.NewSignatureType(nil, nil, nil, types.NewTuple(stringVar), types.NewTuple(intVar), false)},

		{`func($*_) int`, types.NewSignatureType(nil, nil, nil, types.NewTuple(stringVar), types.NewTuple(intVar), false)},
		{`func($*_) int`, types.NewSignatureType(nil, nil, nil, nil, types.NewTuple(intVar), false)},
		{`func($*_) $_`, types.NewSignatureType(nil, nil, nil, nil, types.NewTuple(intVar), false)},

		{`func($t, $t)`, types.NewSignatureType(nil, nil, nil, types.NewTuple(stringVar, stringVar), nil, false)},
		{`func($t, $t)`, types.NewSignatureType(nil, nil, nil, types.NewTuple(intVar, intVar), nil, false)},

		// Any func.
		{`func($*_) $*_`, types.NewSignatureType(nil, nil, nil, nil, nil, false)},
		{`func($*_) $*_`, types.NewSignatureType(nil, nil, nil, types.NewTuple(stringVar, stringVar), nil, false)},
		{`func($*_) $*_`, types.NewSignatureType(nil, nil, nil, types.NewTuple(stringVar), types.NewTuple(intVar), false)},

		{`struct{}`, typeEstruct},
		{`struct{int}`, types.NewStruct([]*types.Var{intVar}, nil)},
		{`struct{string; int}`, types.NewStruct([]*types.Var{stringVar, intVar}, nil)},
		{`struct{$_; string}`, types.NewStruct([]*types.Var{stringVar, stringVar}, nil)},
		{`struct{$_; $_}`, types.NewStruct([]*types.Var{stringVar, intVar}, nil)},
		{`struct{$x; $x}`, types.NewStruct([]*types.Var{intVar, intVar}, nil)},

		// Any struct.
		{`struct{$*_}`, typeEstruct},
		{`struct{$*_}`, structType(intVar, intVar)},

		// Struct has suffix.
		{`struct{$*_; int}`, structType(intVar)},
		{`struct{$*_; int}`, structType(stringVar, stringVar, intVar)},

		// Struct has prefix.
		{`struct{int; $*_}`, structType(intVar)},
		{`struct{int; $*_}`, structType(intVar, stringVar, stringVar)},

		// Struct contains.
		{`struct{$*_; int; $*_}`, structType(intVar)},
		{`struct{$*_; int; $*_}`, structType(stringVar, intVar)},
		{`struct{$*_; int; $*_}`, structType(intVar, stringVar)},
		{`struct{$*_; int; $*_}`, structType(stringVar, intVar, stringVar)},

		// Struct with dups.
		{`struct{$*_; $x; $*_; $x; $*_}`, structType(intVar, intVar)},
		{`struct{$*_; $x; $*_; $x; $*_}`, structType(intVar, intVar, stringVar)},
		{`struct{$*_; $x; $*_; $x; $*_}`, structType(intVar, int32Var, intVar, stringVar)},
		{`struct{$*_; $x; $*_; $x; $*_}`, structType(intVar, int32Var, stringVar, intVar)},
	}

	state := NewMatcherState()
	for _, test := range tests {
		pat, err := Parse(testContext, test.expr)
		if err != nil {
			t.Errorf("parse('%s'): %v", test.expr, err)
			continue
		}
		if !pat.MatchIdentical(state, test.typ) {
			t.Errorf("identical('%s', %s): expected a match",
				test.expr, test.typ.String())
		}
	}
}

func TestIdenticalNegative(t *testing.T) {
	tests := []struct {
		expr string
		typ  types.Type
	}{
		{`int`, typeString},
		{`[]int`, types.NewSlice(typeString)},
		{`[][]int`, types.NewSlice(types.NewSlice(typeString))},
		{`[][]int`, types.NewSlice(typeInt)},
		{`[10]int`, types.NewArray(typeInt, 11)},
		{`[10]int`, types.NewArray(typeString, 10)},
		{`map[int]int`, types.NewMap(typeString, typeString)},
		{`map[int]int`, types.NewMap(typeString, typeInt)},
		{`map[int]int`, types.NewMap(typeInt, typeString)},

		{`unsafe.Pointer`, typeInt},
		{`unsafe.Pointer`, types.NewPointer(typeInt)},
		{`[]unsafe.Pointer`, types.NewSlice(typeInt)},

		{`interface{}`, typeInt},
		{`interface{ $*_ }`, typeString},
		{`interface{ $*_ }`, types.NewArray(typeString, 10)},

		{`*$t`, typeInt},
		{`map[$t]$t`, types.NewMap(typeString, typeInt)},
		{`map[$t]$t`, types.NewMap(typeInt, typeString)},
		{`[$len][$len]int`, types.NewArray(types.NewArray(typeInt, 15), 10)},

		{`chan int`, types.NewChan(types.SendRecv, typeString)},
		{`chan int`, types.NewChan(types.RecvOnly, typeInt)},
		{`chan <- int`, types.NewChan(types.SendRecv, typeInt)},
		{`<- chan int`, types.NewChan(types.SendOnly, typeInt)},

		{`io.Reader`, namedType2("foo/io", "Reader")},
		{`syntax.Regexp`, namedType2("regexp2/syntax", "Regexp")},
		{`syntax.Regexp`, namedType2("regexp2/syntax", "Blah")},

		{`func(int)`, types.NewSignatureType(nil, nil, nil, nil, nil, false)},
		{`func() int`, types.NewSignatureType(nil, nil, nil, types.NewTuple(intVar), nil, false)},
		{`func(int, int)`, types.NewSignatureType(nil, nil, nil, types.NewTuple(intVar, stringVar), nil, false)},
		{`func() string`, types.NewSignatureType(nil, nil, nil, nil, types.NewTuple(intVar), false)},
		{`func(string, string) int`, types.NewSignatureType(nil, nil, nil, types.NewTuple(stringVar), types.NewTuple(intVar), false)},
		{`func(string) string`, types.NewSignatureType(nil, nil, nil, nil, types.NewTuple(stringVar, intVar), false)},

		{`func($_) int`, types.NewSignatureType(nil, nil, nil, types.NewTuple(intVar), types.NewTuple(stringVar), false)},

		{`func($t, $t)`, types.NewSignatureType(nil, nil, nil, types.NewTuple(intVar, stringVar), nil, false)},
		{`func($t, $t)`, types.NewSignatureType(nil, nil, nil, types.NewTuple(stringVar, intVar), nil, false)},

		{`func($*_) int`, types.NewSignatureType(nil, nil, nil, types.NewTuple(stringVar), types.NewTuple(stringVar), false)},
		{`func($*_) int`, types.NewSignatureType(nil, nil, nil, nil, nil, false)},
		{`func($*_) $_`, types.NewSignatureType(nil, nil, nil, nil, nil, false)},

		// Any func negative.
		{`func($*_) $*_`, typeInt},
		{`func($*_) $*_`, types.NewTuple(intVar)},

		{`struct{}`, typeInt},
		{`struct{}`, types.NewStruct([]*types.Var{intVar}, nil)},
		{`struct{int}`, typeEstruct},
		{`struct{int}`, types.NewStruct([]*types.Var{stringVar}, nil)},
		{`struct{string; int}`, types.NewStruct([]*types.Var{intVar, stringVar}, nil)},
		{`struct{$_; string}`, types.NewStruct([]*types.Var{stringVar, stringVar, intVar}, nil)},
		{`struct{$_; $_}`, types.NewStruct([]*types.Var{stringVar}, nil)},
		{`struct{$x; $x}`, types.NewStruct([]*types.Var{intVar, stringVar}, nil)},

		// Any struct negative.
		{`struct{$*_}`, typeInt},

		// Struct has suffix negative.
		{`struct{$*_; int}`, typeEstruct},
		{`struct{$*_; int}`, structType(stringVar)},

		// Struct has prefix negative.
		{`struct{int; $*_}`, typeEstruct},
		{`struct{int; $*_}`, structType(stringVar)},

		// Struct contains negative.
		{`struct{$*_; int; $*_}`, typeEstruct},
		{`struct{$*_; int; $*_}`, structType(stringVar)},
		{`struct{$*_; int; $*_}`, structType(stringVar, int32Var)},

		// Struct with dups negative.
		{`struct{$*_; $x; $*_; $x; $*_}`, typeEstruct},
		{`struct{$*_; $x; $*_; $x; $*_}`, structType(int32Var, intVar)},
		{`struct{$*_; $x; $*_; $x; $*_}`, structType(intVar, int32Var, stringVar)},
		{`struct{$*_; $x; $*_; $x; $*_}`, structType(intVar, int32Var, estructVar, stringVar)},

		// TODO: this should fail as $* is named.
		// We don't support named $* now, but they should be supported.
		// {`struct{$*x; int; $*x}`, structType(stringVar, intVar, intVar)},
	}

	state := NewMatcherState()
	for _, test := range tests {
		pat, err := Parse(testContext, test.expr)
		if err != nil {
			t.Errorf("parse('%s'): %v", test.expr, err)
			continue
		}
		if pat.MatchIdentical(state, test.typ) {
			t.Errorf("identical('%s', %s): unexpected match",
				test.expr, test.typ.String())
		}
	}
}
