package evaltest

// This package is used for quasigo testing.

type Foo struct {
	Prefix string
}

func (*Foo) Method1(x int) string { return "" }
