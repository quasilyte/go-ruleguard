package quickfix

func rangeRuneSlice(s string) {
	for _, ch := range []rune(s) { // want `\Qsuggestion: range s`
		println(ch)
	}

	{
		var ch rune
		for _, ch = range []rune(s[:]) { // want `\Qsuggestion: range s[:]`
			println(ch)
		}
	}

	{
		var ch rune
		for _, ch = range []rune(getString()) { // want `\Qsuggestion: range getString()`
			println(ch)
		}
	}

	{
		for _, ch1 := range []rune("foo") { // want `\Qsuggestion: range "foo"`
			for _, ch2 := range []rune("123") { // want `\Qsuggestion: range "123"`
				println(ch1, ch2)
			}
		}
	}
}

func getString() string {
	return "123"
}
