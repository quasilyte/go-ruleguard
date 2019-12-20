package typematch

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strconv"
	"strings"
)

type patternOp int

const (
	opBuiltinType patternOp = iota
	opPointer
	opVar
	opSlice
	opArray
	opMap
	opChan
	opNamed
)

type Pattern struct {
	typeMatches  map[string]types.Type
	int64Matches map[string]int64

	root *pattern
}

type pattern struct {
	value interface{}
	op    patternOp
	subs  []*pattern
}

type Context struct {
	Imports map[string]string
}

func Parse(ctx *Context, s string) (*Pattern, error) {
	noDollars := strings.ReplaceAll(s, "$", "__")
	n, err := parser.ParseExpr(noDollars)
	if err != nil {
		return nil, err
	}
	root := parseExpr(ctx, n)
	if root == nil {
		return nil, fmt.Errorf("can't convert %s type expression", s)
	}
	p := &Pattern{
		typeMatches:  map[string]types.Type{},
		int64Matches: map[string]int64{},
		root:         root,
	}
	return p, nil
}

var (
	builtinTypeByName = map[string]types.Type{
		"bool":       types.Typ[types.Bool],
		"int":        types.Typ[types.Int],
		"int8":       types.Typ[types.Int8],
		"int16":      types.Typ[types.Int16],
		"int32":      types.Typ[types.Int32],
		"int64":      types.Typ[types.Int64],
		"uint":       types.Typ[types.Uint],
		"uint8":      types.Typ[types.Uint8],
		"uint16":     types.Typ[types.Uint16],
		"uint32":     types.Typ[types.Uint32],
		"uint64":     types.Typ[types.Uint64],
		"uintptr":    types.Typ[types.Uintptr],
		"float32":    types.Typ[types.Float32],
		"float64":    types.Typ[types.Float64],
		"complex64":  types.Typ[types.Complex64],
		"complex128": types.Typ[types.Complex128],
		"string":     types.Typ[types.String],

		"error": types.Universe.Lookup("error").Type(),
	}

	efaceType = types.NewInterfaceType(nil, nil)
)

func parseExpr(ctx *Context, e ast.Expr) *pattern {
	switch e := e.(type) {
	case *ast.Ident:
		basic, ok := builtinTypeByName[e.Name]
		if ok {
			return &pattern{op: opBuiltinType, value: basic}
		}
		if strings.HasPrefix(e.Name, "__") {
			name := strings.TrimPrefix(e.Name, "__")
			return &pattern{op: opVar, value: name}
		}

	case *ast.SelectorExpr:
		pkg, ok := e.X.(*ast.Ident)
		if !ok {
			return nil
		}
		pkgPath, ok := ctx.Imports[pkg.Name]
		if !ok {
			pkgPath = stdlib[pkg.Name]
			if pkgPath == "" {
				return nil
			}
		}
		return &pattern{op: opNamed, value: [2]string{pkgPath, e.Sel.Name}}

	case *ast.StarExpr:
		elem := parseExpr(ctx, e.X)
		if elem == nil {
			return nil
		}
		return &pattern{op: opPointer, subs: []*pattern{elem}}

	case *ast.ArrayType:
		elem := parseExpr(ctx, e.Elt)
		if elem == nil {
			return nil
		}
		if e.Len == nil {
			return &pattern{
				op:   opSlice,
				subs: []*pattern{elem},
			}
		}
		if id, ok := e.Len.(*ast.Ident); ok && strings.HasPrefix(id.Name, "__") {
			name := strings.TrimPrefix(id.Name, "__")
			return &pattern{
				op:    opArray,
				value: name,
				subs:  []*pattern{elem},
			}
		}
		lit, ok := e.Len.(*ast.BasicLit)
		if !ok || lit.Kind != token.INT {
			return nil
		}
		length, err := strconv.ParseInt(lit.Value, 10, 64)
		if err != nil {
			return nil
		}
		return &pattern{
			op:    opArray,
			value: length,
			subs:  []*pattern{elem},
		}

	case *ast.MapType:
		keyType := parseExpr(ctx, e.Key)
		if keyType == nil {
			return nil
		}
		valType := parseExpr(ctx, e.Value)
		if valType == nil {
			return nil
		}
		return &pattern{
			op:   opMap,
			subs: []*pattern{keyType, valType},
		}

	case *ast.ChanType:
		valType := parseExpr(ctx, e.Value)
		if valType == nil {
			return nil
		}
		var dir types.ChanDir
		switch {
		case e.Dir&ast.SEND != 0 && e.Dir&ast.RECV != 0:
			dir = types.SendRecv
		case e.Dir&ast.SEND != 0:
			dir = types.SendOnly
		case e.Dir&ast.RECV != 0:
			dir = types.RecvOnly
		default:
			return nil
		}
		return &pattern{
			op:    opChan,
			value: dir,
			subs:  []*pattern{valType},
		}

	case *ast.ParenExpr:
		return parseExpr(ctx, e.X)

	case *ast.InterfaceType:
		if len(e.Methods.List) == 0 {
			return &pattern{op: opBuiltinType, value: efaceType}
		}
	}

	return nil
}

func (p *Pattern) MatchIdentical(typ types.Type) bool {
	p.reset()
	return p.matchIdentical(p.root, typ)
}

func (p *Pattern) reset() {
	if len(p.int64Matches) != 0 {
		p.int64Matches = map[string]int64{}
	}
	if len(p.typeMatches) != 0 {
		p.typeMatches = map[string]types.Type{}
	}
}

func (p *Pattern) matchIdentical(sub *pattern, typ types.Type) bool {
	switch sub.op {
	case opVar:
		name := sub.value.(string)
		if name == "_" {
			return true
		}
		y, ok := p.typeMatches[name]
		if !ok {
			p.typeMatches[name] = typ
			return true
		}
		if y == nil {
			return typ == nil
		}
		return types.Identical(typ, y)

	case opBuiltinType:
		return types.Identical(typ, sub.value.(types.Type))

	case opPointer:
		typ, ok := typ.(*types.Pointer)
		if !ok {
			return false
		}
		return p.matchIdentical(sub.subs[0], typ.Elem())

	case opSlice:
		typ, ok := typ.(*types.Slice)
		if !ok {
			return false
		}
		return p.matchIdentical(sub.subs[0], typ.Elem())

	case opArray:
		typ, ok := typ.(*types.Array)
		if !ok {
			return false
		}
		var wantLen int64
		switch v := sub.value.(type) {
		case string:
			if v == "_" {
				wantLen = typ.Len()
				break
			}
			length, ok := p.int64Matches[v]
			if ok {
				wantLen = length
			} else {
				p.int64Matches[v] = typ.Len()
				wantLen = typ.Len()
			}
		case int64:
			wantLen = v
		}
		return wantLen == typ.Len() && p.matchIdentical(sub.subs[0], typ.Elem())

	case opMap:
		typ, ok := typ.(*types.Map)
		if !ok {
			return false
		}
		return p.matchIdentical(sub.subs[0], typ.Key()) &&
			p.matchIdentical(sub.subs[1], typ.Elem())

	case opChan:
		typ, ok := typ.(*types.Chan)
		if !ok {
			return false
		}
		dir := sub.value.(types.ChanDir)
		return dir == typ.Dir() && p.matchIdentical(sub.subs[0], typ.Elem())

	case opNamed:
		typ, ok := typ.(*types.Named)
		if !ok {
			return false
		}
		pkgPath := sub.value.([2]string)[0]
		typeName := sub.value.([2]string)[1]
		obj := typ.Obj()
		return obj.Pkg().Path() == pkgPath && typeName == obj.Name()

	default:
		return false
	}
}

var stdlib = map[string]string{
	"adler32":         "hash/adler32",
	"aes":             "crypto/aes",
	"ascii85":         "encoding/ascii85",
	"asn1":            "encoding/asn1",
	"ast":             "go/ast",
	"atomic":          "sync/atomic",
	"base32":          "encoding/base32",
	"base64":          "encoding/base64",
	"big":             "math/big",
	"binary":          "encoding/binary",
	"bits":            "math/bits",
	"bufio":           "bufio",
	"build":           "go/build",
	"bytes":           "bytes",
	"bzip2":           "compress/bzip2",
	"cgi":             "net/http/cgi",
	"cgo":             "runtime/cgo",
	"cipher":          "crypto/cipher",
	"cmplx":           "math/cmplx",
	"color":           "image/color",
	"constant":        "go/constant",
	"context":         "context",
	"cookiejar":       "net/http/cookiejar",
	"crc32":           "hash/crc32",
	"crc64":           "hash/crc64",
	"crypto":          "crypto",
	"csv":             "encoding/csv",
	"debug":           "runtime/debug",
	"des":             "crypto/des",
	"doc":             "go/doc",
	"draw":            "image/draw",
	"driver":          "database/sql/driver",
	"dsa":             "crypto/dsa",
	"dwarf":           "debug/dwarf",
	"ecdsa":           "crypto/ecdsa",
	"ed25519":         "crypto/ed25519",
	"elf":             "debug/elf",
	"elliptic":        "crypto/elliptic",
	"encoding":        "encoding",
	"errors":          "errors",
	"exec":            "os/exec",
	"expvar":          "expvar",
	"fcgi":            "net/http/fcgi",
	"filepath":        "path/filepath",
	"flag":            "flag",
	"flate":           "compress/flate",
	"fmt":             "fmt",
	"fnv":             "hash/fnv",
	"format":          "go/format",
	"gif":             "image/gif",
	"gob":             "encoding/gob",
	"gosym":           "debug/gosym",
	"gzip":            "compress/gzip",
	"hash":            "hash",
	"heap":            "container/heap",
	"hex":             "encoding/hex",
	"hmac":            "crypto/hmac",
	"html":            "html",
	"http":            "net/http",
	"httptest":        "net/http/httptest",
	"httptrace":       "net/http/httptrace",
	"httputil":        "net/http/httputil",
	"image":           "image",
	"importer":        "go/importer",
	"io":              "io",
	"iotest":          "testing/iotest",
	"ioutil":          "io/ioutil",
	"jpeg":            "image/jpeg",
	"json":            "encoding/json",
	"jsonrpc":         "net/rpc/jsonrpc",
	"list":            "container/list",
	"log":             "log",
	"lzw":             "compress/lzw",
	"macho":           "debug/macho",
	"mail":            "net/mail",
	"math":            "math",
	"md5":             "crypto/md5",
	"mime":            "mime",
	"multipart":       "mime/multipart",
	"net":             "net",
	"os":              "os",
	"palette":         "image/color/palette",
	"parse":           "text/template/parse",
	"parser":          "go/parser",
	"path":            "path",
	"pe":              "debug/pe",
	"pem":             "encoding/pem",
	"pkix":            "crypto/x509/pkix",
	"plan9obj":        "debug/plan9obj",
	"plugin":          "plugin",
	"png":             "image/png",
	"pprof":           "runtime/pprof",
	"printer":         "go/printer",
	"quick":           "testing/quick",
	"quotedprintable": "mime/quotedprintable",
	"race":            "runtime/race",
	"rand":            "math/rand",
	"rc4":             "crypto/rc4",
	"reflect":         "reflect",
	"regexp":          "regexp",
	"ring":            "container/ring",
	"rpc":             "net/rpc",
	"rsa":             "crypto/rsa",
	"runtime":         "runtime",
	"scanner":         "text/scanner",
	"sha1":            "crypto/sha1",
	"sha256":          "crypto/sha256",
	"sha512":          "crypto/sha512",
	"signal":          "os/signal",
	"smtp":            "net/smtp",
	"sort":            "sort",
	"sql":             "database/sql",
	"strconv":         "strconv",
	"strings":         "strings",
	"subtle":          "crypto/subtle",
	"suffixarray":     "index/suffixarray",
	"sync":            "sync",
	"syntax":          "regexp/syntax",
	"syscall":         "syscall",
	"syslog":          "log/syslog",
	"tabwriter":       "text/tabwriter",
	"tar":             "archive/tar",
	"template":        "text/template",
	"testing":         "testing",
	"textproto":       "net/textproto",
	"time":            "time",
	"tls":             "crypto/tls",
	"token":           "go/token",
	"trace":           "runtime/trace",
	"types":           "go/types",
	"unicode":         "unicode",
	"unsafe":          "unsafe",
	"url":             "net/url",
	"user":            "os/user",
	"utf16":           "unicode/utf16",
	"utf8":            "unicode/utf8",
	"x509":            "crypto/x509",
	"xml":             "encoding/xml",
	"zip":             "archive/zip",
	"zlib":            "compress/zlib",
}
