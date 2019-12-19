package extra

import "fmt"

func testFormatInt() {
	{
		x16 := int16(342)
		_ = fmt.Sprintf("%d", x16) // want `use strconv.FormatInt\(int64\(x16\), 10\)`
	}
	{
		x64 := int64(32)
		_ = fmt.Sprintf("%d", x64) // want `use strconv.FormatInt\(x64, 10\)`
	}
	{
		// Check that convertibleTo(int64) condition works and rejects this.
		s := struct{}{}
		_ = fmt.Sprintf("%d", s)
	}
}

func testFormatBool() {
	{
		i := int64(4)
		_ = fmt.Sprintf("%t", (i+i)&1 == 0) // want `use strconv.FormatBool\(\(i \+ i\)&1 == 0\)`
	}
}
