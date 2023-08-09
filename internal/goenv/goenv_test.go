package goenv

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		lines  []string
		goroot string
		gopath string
		err    bool
	}{
		// handle windows line-endings
		{
			lines: []string{
				"C:\\Program Files\\Go\r\n",
				"C:\\Users\\me\\go\r\n",
			},
			goroot: "C:\\Program Files\\Go",
			gopath: "C:\\Users\\me\\go",
		},

		// preserve trailing spaces on windows
		{
			lines: []string{
				"C:\\Program Files\\Go \r\n",
				"C:\\Users\\me\\go  \r\n",
			},
			goroot: "C:\\Program Files\\Go ",
			gopath: "C:\\Users\\me\\go  ",
		},

		// handle linux line-endings
		{
			lines: []string{
				"/usr/local/go\n",
				"/home/me/go\n",
			},
			goroot: "/usr/local/go",
			gopath: "/home/me/go",
		},

		// preserve trailing spaces on linux
		{
			lines: []string{
				"/usr/local/go \n",
				"/home/me/go \n",
			},
			goroot: "/usr/local/go ",
			gopath: "/home/me/go ",
		},

		// handle empty value
		{
			lines: []string{
				"\n",
				"/home/me/go\n",
			},
			goroot: "",
			gopath: "/home/me/go",
		},

		// handle short output
		{
			lines: []string{
				"/usr/local/go",
			},
			goroot: "/usr/local/go",
			gopath: "",
		},

		// handle empty output
		{
			lines:  []string{},
			goroot: "",
			gopath: "",
			err:    true,
		},
	}

	for i, test := range tests {
		data := []byte(strings.Join(test.lines, ""))
		vars, err := parseGoEnv([]string{"GOROOT", "GOPATH"}, data)
		if err != nil != test.err {
			t.Fatalf("test %d failed: %v", i, err)
		}
		if vars["GOROOT"] != test.goroot {
			t.Errorf("test %d GOROOT mismatch: have %q, want %q", i, vars["GOROOT"], test.goroot)
			continue
		}
		if vars["GOPATH"] != test.gopath {
			t.Errorf("test %d GOPATH mismatch: have %q, want %q", i, vars["GOPATH"], test.gopath)
			continue
		}
	}
}
