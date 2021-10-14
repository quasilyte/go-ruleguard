package comments

//go:embed

/*foo-bar: baz*/ // want `\Qdirective should be written as //foo-bar`

/*go:noinline*/ // want `\Qdirective should be written as //go`
func f() {
	_ = 12 // nolint // want `\Qremove a space between // and "nolint" directive`

	_ = 30 // nolint2 foo bar // want `\Qsuggestion: //nolint2 foo bar`

	/*
		nolint // want `\Qdon't put "nolint" inside a multi-line comment`
	*/

	//go:baddirective // want `\Qdon't use baddirective go directive`
	//go:noinline
	//go:generate foo bar

	//nolint:gocritic // want `\Qhey, this is kinda upsetting`

	// This is a begining // want `\Q"begining" may contain a typo`
	// Of a bizzare text with typos. // want `\Q"bizzare" may contain a typo`

	// I can't give you a buisness advice. // want `\Q"buisness advice" may contain a typo`

	// calender // want `\Qfirst=calender`
	// cemetary // want `\Qsecond=cemetary`

	// collegue // want `\Qx="collegue"`
	// commitee // want `\Qx=""`
}
