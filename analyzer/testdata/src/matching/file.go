package golint

func sink(args ...interface{}) {}

func multiexpr() (int, int, int) {
	sink(1, 1)         // want `repeated expression in list`
	_ = []int{0, 1, 1} // want `repeated expression in list`

	_ = []string{
		"",
		"x", // want `repeated expression in list`
		"x",
		"",
		"y", "y", // want `repeated expression in list`
		"",
		"z",
	}

	return 1, 1, 0 // want `repeated expression in list`
}
