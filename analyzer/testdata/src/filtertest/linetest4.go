package filtertest

func _() {
	lineTest("", "line 4") // want `YES`
	lineTest("", "line 4")
}
