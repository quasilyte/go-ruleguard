package gocritic

func suspeciousAppends() {
	var xs []int
	var ys []int

	xs = append(ys, 1) // want `\Qappend result not assigned to the same slice`
	ys = append(xs, 1) // want `\Qappend result not assigned to the same slice`

	{
		xs2 := xs
		xs = append(xs2, 1) // want `\Qappend result not assigned to the same slice`
		xs2 = append(xs, 1) // want `\Qappend result not assigned to the same slice`
	}
}

func normalAppends() {
	var xs, ys []int

	xs = append(xs, 1)
	ys = append(ys, 1, 2)
	xs = append(xs, ys[0], xs[0])
}

func permittedAppends() {
	var xs, ys []int

	// We're trying to detect `x = append(y, ...)` patterns
	// where y is used instead of x by mistake, so lines below
	// do not trigger a warning.

	xs0 := append(xs, 1)
	xs1 := append(xs, 1)
	ys0 := append(ys, 1)

	// Also permit to assign to "_".
	_ = append(xs, xs0[0], xs1[1], ys0[0])

	{
		var m map[int][]int
		xs := m[0]
		m[0] = append(xs, 1)
	}

	// Sliced xs is still xs.
	xs = append(xs[:0], 1)
	xs = append(xs[1:], 2)

	// OK to use slice literals.
	xs = append([]int{}, 1)
	xs = append([]int{1, 2}, 1)

	// Also OK to use slices returned by a function calls.
	xs = append(*new([]int), 1)
	*(new([]int)) = append(*(new([]int)), 1)

	// This prepends ys to the xs. Common idiom.
	xs = append(ys, xs...)
	xs = append(ys, xs[1:]...)

	// Scratch array idiom.
	var scratch [10]int
	xs = append(scratch[:], 1)
	xs = append(scratch[1:5], 2)

	var withSlices struct {
		a []int
		b []int
	}
	withSlices.a = append(withSlices.a, 1)
	withSlices.b = append(withSlices.b, 1)

	var xsMap map[string][]int
	xsMap["10"] = append(xsMap["10"], 1, 2)
}

func appendNotInAssignment() {
	var xs, ys []int

	// These are somewhat weird, but has nothing
	// to do with diagnostic this checker wants to perform.

	var v1 = append(xs, 1)
	var (
		v2 = append(xs, v1[0])
		v3 = append(v2[1:], ys[0])
	)
	v4 := append(v3, xs[0])
	{
		v3 := append(v4, 1)
		_ = v3
	}
}
