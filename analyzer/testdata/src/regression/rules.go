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
	m.Match(`println($x)`).
		Where(!(m["x"].Const && m["x"].Type.Is("int"))).
		Report("$x is not a constexpr int")
}

func issue192(m dsl.Matcher) {
	m.Match(`fmt.Print(fmt.Sprintf($format, $*args))`).
		Suggest(`fmt.Printf($format, $args)`)

	m.Match(`fmt.Println(fmt.Sprintf($format, $*args, $last))`).
		Suggest(`fmt.Printf($format+"\n", $args, $last)`)
}
