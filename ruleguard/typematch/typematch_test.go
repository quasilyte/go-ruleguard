package typematch

import (
	"go/types"
	"testing"
)

var (
	typeInt    = types.Typ[types.Int]
	typeString = types.Typ[types.String]
)

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
	}

	for _, test := range tests {
		pat, err := Parse(test.expr)
		if err != nil {
			t.Errorf("parse('%s'): %v", test.expr, err)
			continue
		}
		if !pat.MatchIdentical(test.typ) {
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
		{`interface{}`, typeInt},

		{`*$t`, typeInt},
		{`map[$t]$t`, types.NewMap(typeString, typeInt)},
		{`map[$t]$t`, types.NewMap(typeInt, typeString)},
		{`[$len][$len]int`, types.NewArray(types.NewArray(typeInt, 15), 10)},
	}

	for _, test := range tests {
		pat, err := Parse(test.expr)
		if err != nil {
			t.Errorf("parse('%s'): %v", test.expr, err)
			continue
		}
		if pat.MatchIdentical(test.typ) {
			t.Errorf("identical('%s', %s): unexpected match",
				test.expr, test.typ.String())
		}
	}
}
