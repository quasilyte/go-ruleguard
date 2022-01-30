package regression

func _() {
	_ = map[string]int{}       // want `\Qcreating a map`
	_ = make(map[int][]string) // want `\Qcreating a map`
}
