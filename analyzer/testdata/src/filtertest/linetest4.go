package filtertest

func _() {
	lineTest("", "line 4") // want `true`
	lineTest("", "line 4")
}
