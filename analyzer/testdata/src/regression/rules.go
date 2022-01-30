//go:build ignore
// +build ignore

package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
)

func issue68(m dsl.Matcher) {
	m.Match(`func $_($_ *testing.T) { $p; $*_ }`).Where(m["p"].Text != "t.Parallel()").Report(`Not a parallel test`)
	m.Match(`func $_($_ *testing.T) { $p; $*_ }`).Where(m["p"].Text == "t.Parallel()").Report(`Parallel test`)
}

func issue72(m dsl.Matcher) {
	m.Match("fmt.Sprintf(`\"%s\" <%s>`, $name, $email)",
		"fmt.Sprintf(`\"%s\"<%s>`, $name, $email)",
		`fmt.Sprintf("\"%s\" <%s>", $name, $email)`,
		`fmt.Sprintf("\"%s\"<%s>", $name, $email)`,
		`fmt.Sprintf("%s<%s>", $name, $email)`).
		Report("use net/mail Address.String() instead of fmt.Sprintf()")
}

func issue115(m dsl.Matcher) {
	m.Match(`println($x, "!constexpr int")`).
		Where(!(m["x"].Const && m["x"].Type.Is("int"))).
		Report("$x is not a constexpr int")
}

func issue192(m dsl.Matcher) {
	m.Match(`fmt.Print(fmt.Sprintf($format, $*args))`).
		Suggest(`fmt.Printf($format, $args)`)

	m.Match(`fmt.Println(fmt.Sprintf($format, $*args, $last))`).
		Suggest(`fmt.Printf($format+"\n", $args, $last)`)
}

func issue291(m dsl.Matcher) {
	m.Match(`const ( $_ = $iota; $*_ )`).
		Where(m["iota"].Text == "iota").
		At(m["iota"]).
		Report("avoid use of iota without explicit type")

	m.Match(`const ( $_ $_ = $iota; $*_ )`).
		Where(m["iota"].Text == "iota").
		At(m["iota"]).
		Report("good, have explicit type")
}

func issue339(m dsl.Matcher) {
	m.Match(`println("339"); println("x")`).Report("pattern1")
	m.Match(`println("x"); println("339")`).Report("pattern2")
}

func issue315(m dsl.Matcher) {
	m.Match(
		`func $name($*_) $arg { $*_ }`,
		`func $name($*_) ($arg, $_) { $*_ }`,
		`func $name($*_) ($_, $arg, $_) { $*_ }`,
		`func ($_ $_) $name($*_) ($arg, $_) { $*_ }`,
	).Where(
		m["name"].Text.Matches(`^[A-Z]`) &&
			m["arg"].Type.Underlying().Is(`interface{ $*_ }`) &&
			!m["arg"].Type.Is(`error`),
	).Report(`return concrete type instead of $arg`).At(m["name"])
}

func issue360(m dsl.Matcher) {
	m.Match(`$_{$*_, $_: strings.Compare($s1, $_), $*_}`,
		`$_{$*_, strings.Compare($s1, $_): $_, $*_}`).
		Report(`don't use strings.Compare`).
		At(m["s1"])
}

func issue372(m dsl.Matcher) {
	m.Match("$x{}", "make($x)").
		Where(m["x"].Type.Is("map[$k]$v")).
		Report("creating a map")
}
