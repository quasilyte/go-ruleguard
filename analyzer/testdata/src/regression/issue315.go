package regression

import (
	"io"
)

type Issue315_Iface interface {
	Example()
}

func Issue315_Func() {
}

func Issue315_FuncErr() error {
	return nil
}

func Issue315_Func1() io.Reader { // want `\Qreturn concrete type instead of io.Reader`
	return nil
}

func Issue315_Func2() (io.Reader, error) { // want `\Qreturn concrete type instead of io.Reader`
	return nil, nil
}

func Issue315_Func3() (int, Issue315_Iface, error) { // want `\Qreturn concrete type instead of Issue315_Iface`
	return 0, nil, nil
}

type Issue315_Example struct{}

func (example Issue315_Example) Func1(x int) (Issue315_Iface, error) { // want `\Qreturn concrete type instead of Issue315_Iface`
	return nil, nil
}
