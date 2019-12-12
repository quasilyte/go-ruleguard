package filtertest

func detectType() {
	var i1, i2 int
	var s1, s2 string
	typeTest(s1 + s2) // want `information: concat`
	typeTest(i1 + i2) // want `information: addition`
	typeTest(s1 > s2) // want `information: s1 is !int`
	typeTest(i1 > i2) // want `information: i1 is !string and is pure`
	typeTest(random() > i2)
}

func detectPure(x int) {
	pureTest(random()) // want `information: !pure`
	pureTest(x * x)    // want `information: pure`
}
