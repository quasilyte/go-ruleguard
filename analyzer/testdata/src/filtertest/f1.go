package filtertest

func detectType() {
	var i1, i2 int
	var ii []int
	var s1, s2 string
	var ss []string
	typeTest(s1 + s2) // want `concat`
	typeTest(i1 + i2) // want `addition`
	typeTest(s1 > s2) // want `s1 !is\(int\)`
	typeTest(i1 > i2) // want `i1 !is\(string\) && pure`
	typeTest(random() > i2)
	typeTest(ss, ss) // want `ss is\(\[\]string\)`
	typeTest(ii, ii)
}

func detectPure(x int) {
	pureTest(random()) // want `!pure`
	pureTest(x * x)    // want `pure`
}
