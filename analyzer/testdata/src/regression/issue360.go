package regression

import "strings"

func _(s1, s2 string) {
	_ = map[int]int{
		strings.Compare("", ""): 0, // want `\Qdon't use strings.Compare`
	}

	_ = map[int]int{
		10: strings.Compare("", ""), // want `\Qdon't use strings.Compare`
	}

	_ = map[int]string{
		10:                      "a",
		strings.Compare(s1, s2): s2, // want `\Qdon't use strings.Compare`
		20:                      "b",
	}

	_ = map[int]string{
		10:                      "a",
		20:                      "b",
		strings.Compare(s1, s2): s2, // want `\Qdon't use strings.Compare`
	}

	_ = map[int]string{
		strings.Compare(s1, s2): s2, // want `\Qdon't use strings.Compare`
		10:                      "a",
		20:                      "b",
	}

}
