package main

import "strconv"

func main() {
	s := "16"

	i, err := strconv.Atoi(s)
	println(i)
	println(err == nil)

	i2, err2 := strconv.Atoi("bad")
	println(i2)
	println(err2.Error())

	println(strconv.Itoa(140))
	println(strconv.Itoa(i) == s)

	i, err2 = strconv.Atoi("foo")
	println(i)
	println(err2.Error())

	i, err2 = strconv.Atoi("-349")
	println(i)
	if err2 == nil {
		println("err2 is nil")
	}
}
