// Copyright (c) 2017, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package gogrep

import (
	"bytes"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFiles(t *testing.T) {
	argsList := [][]string{
		{"-x", "foo", "-s", "bar"},
		{"-x", "go func() { $f($*a) }()", "-s", "go $f($*a)"},
	}
	files := []struct{ orig, want string }{
		{
			// TODO: "func foo()" should get changed too
			"package p\n\nfunc foo()\nfunc bar()\nfunc f1() { foo() }\n",
			"package p\n\nfunc foo()\nfunc bar()\nfunc f1() { bar() }\n",
		},
		{
			"// package p doc\npackage p\n\nfunc f2() { foo() }\n",
			"// package p doc\npackage p\n\nfunc f2() { bar() }\n",
		},
		{
			`package p
func fn(int)
func f3() {
	go func() {
		// comment
		fn(0)
	}()
}
`,
			`package p

func fn(int)
func f3() {

	// comment
	go fn(0)

}
`,
		},
	}
	dir, err := ioutil.TempDir("", "gogrep-write")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	var paths []string
	for i, file := range files {
		path := filepath.Join(dir, fmt.Sprintf("f%02d.go", i))
		if err := ioutil.WriteFile(path, []byte(file.orig), 0644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	for _, args := range argsList {
		args = append(args, "-w")
		args = append(args, paths...)

		m := matcher{ctx: &build.Default}
		var buf bytes.Buffer
		m.out = &buf
		if err := m.fromArgs(".", args); err != nil {
			t.Fatalf("didn't want error, but got %q", err)
		}
		gotOut := buf.String()
		if gotOut != "" {
			t.Fatalf("got non-empty output:\n%s", gotOut)
		}
	}

	for i, path := range paths {
		gotBs, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		got := string(gotBs)
		want := files[i].want
		if got != want {
			t.Fatalf("file %d mismatch:\nwant:\n%sgot:\n%s",
				i, want, got)
		}
	}
}
