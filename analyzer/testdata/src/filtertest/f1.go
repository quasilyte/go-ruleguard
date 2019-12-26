package filtertest

type implementsAll struct{}

func (implementsAll) Read([]byte) (int, error) { return 0, nil }
func (implementsAll) String() string           { return "" }

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
	typeTest("2 type filters", i1)
	typeTest("2 type filters", s1)
	typeTest("2 type filters", ii) // want `ii !is\(string\) && !is\(int\)`

	typeTest(implementsAll{}, "implements io.Reader") // want `YES`
	typeTest(i1, "implements io.Reader")
	typeTest(ss, "implements io.Reader")
	typeTest(implementsAll{}, "implements foolib.Stringer") // want `YES`
	typeTest(i1, "implements foolib.Stringer")
	typeTest(ss, "implements foolib.Stringer")

	typeTest([100]byte{}, "size>=100") // want `YES`
	typeTest([105]byte{}, "size>=100") // want `YES`
	typeTest([10]byte{}, "size>=100")
	typeTest([100]byte{}, "size<=100") // want `YES`
	typeTest([105]byte{}, "size<=100")
	typeTest([10]byte{}, "size<=100") // want `YES`
	typeTest([100]byte{}, "size>100")
	typeTest([105]byte{}, "size>100") // want `YES`
	typeTest([10]byte{}, "size>100")
	typeTest([100]byte{}, "size<100")
	typeTest([105]byte{}, "size<100")
	typeTest([10]byte{}, "size<100")   // want `YES`
	typeTest([100]byte{}, "size==100") // want `YES`
	typeTest([105]byte{}, "size==100")
	typeTest([10]byte{}, "size==100")
	typeTest([100]byte{}, "size!=100")
	typeTest([105]byte{}, "size!=100") // want `YES`
	typeTest([10]byte{}, "size!=100")  // want `YES`
}

func detectPure(x int) {
	pureTest(random()) // want `!pure`
	pureTest(x * x)    // want `pure`
}
