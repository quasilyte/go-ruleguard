package goversion

import (
	"io"
	"io/ioutil"
)

func _(r io.Reader) {
	ioutil.ReadAll(r) // want `\Qioutil.ReadAll is deprecated, use io.ReadAll instead`
}

func test(s string) {}

func _() {
	test(">=1.17")
	test(">=1.16") // want `\Qtrue`
	test(">=1.15") // want `\Qtrue`

	test("<=1.17") // want `\Qtrue`
	test("<=1.16") // want `\Qtrue`
	test("<=1.15")

	test(">1.15") // want `\Qtrue`
	test(">1.17")

	test("<1.17") // want `\Qtrue`
	test("<1.16")

	test("<2.0 && >1.0") // want `\Qtrue`
	test("<1.10 && >1.90")

	test("==1.16") // want `\Qtrue`
	test("==1.17")
	test("!=1.16")
}
