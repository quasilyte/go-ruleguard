package suggest

func example(x, y int) {
	_ = (x == 1) || (y == 1) // want `\Qsuggestion: x == 1 || y == 1`
}
