package regression

func testIssue115() {
	intFunc := func() int { return 19 }
	stringFunc := func() string { return "19" }

	println(13, "!constexpr int")
	println(43+5, "!constexpr int")

	println("foo", "!constexpr int")        // want `\Q"foo" is not a constexpr int`
	println(intFunc(), "!constexpr int")    // want `\QintFunc() is not a constexpr int`
	println(stringFunc(), "!constexpr int") // want `\QstringFunc() is not a constexpr int`
}
