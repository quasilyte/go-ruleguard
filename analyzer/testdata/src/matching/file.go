package matching

func sink(args ...interface{}) {}
func expensive()               {}

func multiexpr() (int, int, int) {
	sink(1, 1)         // want `\Qrepeated expression in list`
	_ = []int{0, 1, 1} // want `\Qrepeated expression in list`

	_ = []string{
		"",
		"x", // want `\Qrepeated expression in list`
		"x",
		"",
		"y", "y", // want `\Qrepeated expression in list`
		"",
		"z",
	}

	return 1, 1, 0 // want `\Qrepeated expression in list`
}

func rangeClause() {
	{
		var xs []int
		for _, x := range xs[:] { // want `\Qredundant slicing of a range expression`
			println(x)
		}
		for _, x := range xs {
			println(x)
		}
		for i := range xs[:] { // want `\Qredundant slicing of a range expression`
			println(i)
		}
		for i := range xs {
			println(i)
		}
	}

	{
		var arr [10]int
		for _, x := range arr[:] {
			println(x)
		}
		for _, x := range arr {
			println(x)
		}
		for i := range arr[:] {
			println(i)
		}
		for i := range arr {
			println(i)
		}
	}
}

func submatchContains() {
	{
		type Book struct {
			AuthorID int
		}
		var books []Book
		m := make(map[int][]Book)
		for _, b := range books {
			m[b.AuthorID] = append(m[b.AuthorID], b) // want `\Qm[b.AuthorID] contains b`
		}
	}

	{
		var b1 []byte
		var b2 []byte
		copy(b1, b2)
		copy(b1[:], b2)    // want `\Qcopy() contains a slicing operation`
		copy(b1, b2[:])    // want `\Qcopy() contains a slicing operation`
		copy(b1[:], b2[:]) // want `\Qcopy() contains a slicing operation`
	}

	_ = func() error {
		var err error
		sink(err) // want `\Qsink(err) call not followed by return err`
		return nil
	}
	_ = func() error {
		var err error
		sink(err)
		return err
	}
	_ = func() error {
		var err2 error
		sink(err2) // want `\Qsink(err2) call not followed by return err2`
		return nil
	}
	_ = func() error {
		var err2 error
		sink(err2)
		return err2
	}

	for { // want `\Qexpensive call inside a loop`
		expensive()
	}
	var cond bool
	for { // want `\Qexpensive call inside a loop`
		if cond {
			expensive()
		}
	}
	for {
		if cond {
		}
	}
	for { // want `\Qexpensive call inside a loop`
		for { // want `\Qexpensive call inside a loop`
			expensive()
		}
	}

	{
		type object struct {
			val int
		}
		var objects []object
		if objects != nil && objects[0].val != 0 { // want `\Qnil check may not be enough to access objects[0], check for len`
		}
	}
}
