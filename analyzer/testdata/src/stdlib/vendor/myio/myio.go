package myio

import "io"

type Writer interface {
	io.Writer
}

func WriteString(w Writer, s string) {}
