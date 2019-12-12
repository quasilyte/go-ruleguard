// Copyright (c) 2017, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package gogrep

import (
	"bytes"
	"fmt"
	"go/build"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	ctx := build.Default
	baseDir, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatal(err)
	}
	m := matcher{ctx: &ctx}
	tests := []struct {
		args []string
		want interface{}
	}{
		{
			[]string{"-x", "var _ = $x", "two/file1.go", "two/file2.go"},
			`
				two/file1.go:3:1: var _ = "file1"
				two/file2.go:3:1: var _ = "file2"
			`,
		},
		// TODO(mvdan): reenable once
		// https://github.com/golang/go/issues/29280 is fixed
		// {
		// 	[]string{"-x", "var _ = $x", "noexist.go"},
		// 	fmt.Errorf("packages not found"),
		// },
		// {
		// 	[]string{"-x", "var _ = $x", "-x", "$x", "-a", "type(string)", "noexist.go"},
		// 	fmt.Errorf("packages not found"),
		// },
		{
			[]string{"-x", "var _ = $x", "./p1"},
			`p1/file1.go:3:1: var _ = "file1"`,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "type(string)", "-p", "2", "./p1"},
			`p1/file1.go:3:1: var _ = "file1"`,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "type(int)", "./p1"},
			``, // different type
		},
		{
			[]string{"-x", "var _ = $x", "./p1/..."},
			`
				p1/file1.go:3:1: var _ = "file1"
				p1/p2/file1.go:3:1: var _ = "file1"
				p1/p2/file2.go:3:1: var _ = "file2"
				p1/p3/testp/file1.go:3:1: var _ = "file1"
				p1/testp/file1.go:3:1: var _ = "file1"
			`,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "type(string)", "-p", "2", "./p1/..."},
			`
				p1/file1.go:3:1: var _ = "file1"
				p1/p2/file1.go:3:1: var _ = "file1"
				p1/p2/file2.go:3:1: var _ = "file2"
				p1/p3/testp/file1.go:3:1: var _ = "file1"
				p1/testp/file1.go:3:1: var _ = "file1"
			`,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "type(string)", "-p", "2", "-r", "./p1"},
			`
				p1/file1.go:3:1: var _ = "file1"
				p1/p2/file1.go:3:1: var _ = "file1"
				p1/p2/file2.go:3:1: var _ = "file2"
			`,
		},
		{
			[]string{"-x", "var _ = $x", "longstr.go"},
			`
				longstr.go:3:1: var _ = ` + "`single line`" + `
				longstr.go:4:1: var _ = "some\nmultiline\nstring"
			`,
		},
		{
			[]string{"-x", "if $_ { $*_ }", "longstmt.go"},
			`longstmt.go:7:2: if true { foo(); bar(); }`,
		},
		{
			[]string{"-x", "1, 2, 3, 4, 5", "exprlist.go"},
			`exprlist.go:5:13: 1, 2, 3, 4, 5`,
		},
	}
	for i, tc := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			var buf bytes.Buffer
			m.out = &buf
			err := m.fromArgs(baseDir, tc.args)
			switch x := tc.want.(type) {
			case error:
				if err == nil {
					t.Fatalf("wanted error %q, got none", x)
				}
				want, got := x.Error(), err.Error()
				want = filepath.FromSlash(want)
				if !strings.Contains(got, want) {
					t.Fatalf("wanted error %q, got %q", want, got)
				}
			case string:
				if err != nil {
					t.Fatalf("didn't want error, but got %q", err)
				}
				want := strings.TrimSpace(strings.Replace(x, "\t", "", -1))
				got := strings.TrimSpace(buf.String())
				want = filepath.FromSlash(want)
				if want != got {
					t.Fatalf("wanted:\n%s\ngot:\n%s", want, got)
				}
			default:
				t.Fatalf("unknown want type %T", x)
			}
		})
	}
}
