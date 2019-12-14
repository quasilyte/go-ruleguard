package filtertest

func detectType() {
	var i1, i2 int
	var ii []int
	var s1, s2 string
	var ss []string
	typeTest(s1 + s2) // want `info: concat`
	typeTest(i1 + i2) // want `info: addition`
	typeTest(s1 > s2) // want `info: s1 !is\(int\)`
	typeTest(i1 > i2) // want `info: i1 !is\(string\) && pure`
	typeTest(random() > i2)
	typeTest(ss, ss) // want `info: ss is\(\[\]string\)`
	typeTest(ii, ii)
}

func detectPure(x int) {
	pureTest(random()) // want `info: !pure`
	pureTest(x * x)    // want `info: pure`
}
