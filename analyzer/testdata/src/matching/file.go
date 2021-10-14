package golint

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
