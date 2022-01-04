package matching

func sink(args ...interface{}) {}

func multiexpr() (int, int, int) {
	sink(1, 1)         // want `\Qrepeated expression in list`
	_ = []int{0, 1, 1} // want `\Qrepeated expression in list`

	_ = []string{
		"",
		"x", // want `\Qrepeated expression in list`
		"x",
		"",
		"y", "y", // want `\Qrepeated expression in list`
		"",
		"z",
	}

	return 1, 1, 0 // want `\Qrepeated expression in list`
}

func rangeClause() {
	{
		var xs []int
		for _, x := range xs[:] { // want `\Qredundant slicing of a range expression`
			println(x)
		}
		for _, x := range xs {
			println(x)
		}
		for i := range xs[:] { // want `\Qredundant slicing of a range expression`
			println(i)
		}
		for i := range xs {
			println(i)
		}
	}

	{
		var arr [10]int
		for _, x := range arr[:] {
			println(x)
		}
		for _, x := range arr {
			println(x)
		}
		for i := range arr[:] {
			println(i)
		}
		for i := range arr {
			println(i)
		}
	}
}
