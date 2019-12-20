package suggest

func example(x, y int) {
	_ = (x == 1) || (y == 1) // want `suggestion: x == 1 || y == 1`
}
