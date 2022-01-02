package regression

func _() {
	println("339") // want `\Qpattern1`
	println("x")   // want `\Qpattern2`

	println("339") // want `\Qpattern1`

	println("x")
}

func _() {
	println("x")   // want `\Qpattern2`
	println("339") // want `\Qpattern1`

	println("x")   // want `\Qpattern2`
	println("339") // want `\Qpattern1`

	println("x")
	println("x") // want `\Qpattern2`
	println("339")
}
