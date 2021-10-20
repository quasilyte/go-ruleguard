package goenv

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		goos   string
		lines  []string
		goroot string
		gopath string
	}{
		{
			goos: "windows",
			lines: []string{
				"set GOROOT=C:\\Program Files\\Go\r\n",
				"set GOPATH=C:\\Users\\me\\go\r\n",
			},
			goroot: "C:\\Program Files\\Go",
			gopath: "C:\\Users\\me\\go",
		},

		// Don't do trim on Windows.
		{
			goos: "windows",
			lines: []string{
				"set GOROOT=C:\\Program Files\\Go \r\n",
				"set GOPATH=C:\\Users\\me\\go  \r\n",
			},
			goroot: "C:\\Program Files\\Go ",
			gopath: "C:\\Users\\me\\go  ",
		},

		{
			goos: "linux",
			lines: []string{
				"GOROOT=\"/usr/local/go\"\n",
				"GOPATH=\"/home/me/go\"\n",
			},
			goroot: "/usr/local/go",
			gopath: "/home/me/go",
		},

		// Trim lines on Linux.
		{
			goos: "linux",
			lines: []string{
				" GOROOT=\"/usr/local/go\"  \n",
				"GOPATH=\"/home/me/go\"  \n",
			},
			goroot: "/usr/local/go",
			gopath: "/home/me/go",
		},

		// Quotes preserve the whitespace.
		{
			goos: "linux",
			lines: []string{
				" GOROOT=\"/usr/local/go \"  \n",
				"GOPATH=\"/home/me/go \"  \n",
			},
			goroot: "/usr/local/go ",
			gopath: "/home/me/go ",
		},
	}

	for i, test := range tests {
		data := []byte(strings.Join(test.lines, ""))
		vars, err := parseGoEnv(data, test.goos)
		if err != nil {
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
