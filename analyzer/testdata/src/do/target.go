package do

func Example() {
	test("custom report")  // want `\QHello, World!`
	test("custom suggest") // want `\Qsuggestion: Hello, World!`

	var x int
	test("var text", "str") // want `\Q"str"`
	test("var text", x+1)   // want `\Qx+1`

	test("trim prefix", "hello, world", "hello")   // want `\Q, world`
	test("trim prefix", "hello, world", "hello, ") // want `\Qworld`
	test("trim prefix", "hello, world", "???")     // want `\Qhello, world`

	test("report empty string", "")        // want `\Qempty string`
	test("report empty string", "example") // want `\Qnon-empty string`

	test("report type", 13)       // want `\Qint`
	test("report type", "str")    // want `\Qstring`
	test("report type", []int{1}) // want `\Q[]int`
	test("report type", x)        // want `\Qint`
	test("report type", &x)       // want `\Q*int`

	test("types identical", 1, 1)   // want `true`
	test("types identical", x, x)   // want `true`
	test("types identical", x, &x)  // want `false`
	test("types identical", 1, 1.5) // want `false`
}

func test(args ...interface{}) {}
