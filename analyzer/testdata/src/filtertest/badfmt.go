package filtertest

func _() {
	_ = ((1))
	_ = (((1))) // want `\Qsuspicious tripple parens`
}
