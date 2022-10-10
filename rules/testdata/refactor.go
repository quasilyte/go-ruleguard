package target

import (
	"sort"
)

func sortFuncs() {
	var ints []int
	var strs []string
	var floats []float64

	sort.Slice(ints, func(i, j int) bool { // want `\QsortFuncs: suggestion: sort.Ints(ints)`
		return ints[i] < ints[j]
	})

	sort.Slice(strs, func(i, j int) bool { // want `\QsortFuncs: suggestion: sort.Strings(strs)`
		return strs[i] < strs[j]
	})

	sort.Slice(floats, func(i, j int) bool { // want `\QsortFuncs: suggestion: sort.Float64s(floats)`
		return floats[i] < floats[j]
	})
}
