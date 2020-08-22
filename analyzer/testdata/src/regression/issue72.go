package regression

import "fmt"

func testIssue72() {
	_ = fmt.Sprintf("%s<%s>", "name", "email@domain.example")      // want `\Quse net/mail Address.String() instead of fmt.Sprintf()`
	_ = fmt.Sprintf("\"%s\" <%s>", "name", "email@domain.example") // want `\Quse net/mail Address.String() instead of fmt.Sprintf()`
	_ = fmt.Sprintf(`"%s" <%s>`, "name", "email@domain.example")   // want `\Quse net/mail Address.String() instead of fmt.Sprintf()`
	_ = fmt.Sprintf(`"%s"<%s>`, "name", "email@domain.example")    // want `\Quse net/mail Address.String() instead of fmt.Sprintf()`
}
