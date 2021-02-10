package regression

import "fmt"

func testIssue192() {
	fmt.Print(fmt.Sprintf("abc"))        // want `\Qfmt.Printf("abc", )`
	fmt.Print(fmt.Sprintf("%d", 10))     // want `\Qfmt.Printf("%d", 10)`
	fmt.Print(fmt.Sprintf("%d%d", 1, 2)) // want `\Qfmt.Printf("%d%d", 1, 2)`

	fmt.Println(fmt.Sprintf("%d", 10))                // want `\Qfmt.Printf("%d"+"\n", , 10)`
	fmt.Println(fmt.Sprintf("%d%d", 10, 20))          // want `\Qfmt.Printf("%d%d"+"\n", 10, 20)`
	fmt.Println(fmt.Sprintf("%d%d%s", 10, 20, "abc")) // want `\Qfmt.Printf("%d%d%s"+"\n", 10, 20, "abc")`
}
