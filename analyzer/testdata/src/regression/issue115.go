package regression

func testIssue115() {
	intFunc := func() int { return 19 }
	stringFunc := func() string { return "19" }

	println(13)
	println(43 + 5)

	println("foo")        // want `\Q"foo" is not a constexpr int`
	println(intFunc())    // want `\QintFunc() is not a constexpr int`
	println(stringFunc()) // want `\QstringFunc() is not a constexpr int`
}
