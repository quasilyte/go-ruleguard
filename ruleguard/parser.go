package ruleguard

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"path"
	"regexp"
	"strconv"

	"github.com/quasilyte/go-ruleguard/internal/mvdan.cc/gogrep"
	"github.com/quasilyte/go-ruleguard/ruleguard/typematch"
)

type parseError string

func (e parseError) Error() string { return string(e) }

type rulesParser struct {
	filename string
	group    string
	fset     *token.FileSet
	res      *GoRuleSet
	types    *types.Info

	itab        *typematch.ImportsTab
	stdImporter types.Importer // TODO(quasilyte): share importer with gogrep?
	srcImporter types.Importer
}

func newRulesParser() *rulesParser {
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

	// TODO(quasilyte): do we need to pass the fileset here?
	fset := token.NewFileSet()
	return &rulesParser{
		itab:        typematch.NewImportsTab(stdlib),
		stdImporter: importer.Default(),
		srcImporter: importer.ForCompiler(fset, "source", nil),
	}
}

func (p *rulesParser) ParseFile(filename string, fset *token.FileSet, r io.Reader) (*GoRuleSet, error) {
	p.filename = filename
	p.fset = fset
	p.res = &GoRuleSet{
		local:     &scopedGoRuleSet{},
		universal: &scopedGoRuleSet{},
	}

	parserFlags := parser.Mode(0)
	f, err := parser.ParseFile(fset, filename, r, parserFlags)
	if err != nil {
		return nil, fmt.Errorf("parse file error: %v", err)
	}

	if f.Name.Name != "gorules" {
		return nil, fmt.Errorf("expected a gorules package name, found %s", f.Name.Name)
	}

	typechecker := types.Config{Importer: p.srcImporter}
	p.types = &types.Info{Types: map[ast.Expr]types.TypeAndValue{}}
	_, err = typechecker.Check("gorules", fset, []*ast.File{f}, p.types)
	if err != nil {
		return nil, fmt.Errorf("typechecker error: %v", err)
	}

	for _, decl := range f.Decls {
		decl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if err := p.parseRuleGroup(decl); err != nil {
			return nil, err
		}
	}

	return p.res, nil
}

func (p *rulesParser) parseRuleGroup(f *ast.FuncDecl) (err error) {
	defer func() {
		rv := recover()
		if rv == nil {
			return
		}
		if parseErr, ok := rv.(parseError); ok {
			err = parseErr
			return
		}
		panic(rv) // not our panic
	}()

	if f.Name.String() == "_" {
		return p.errorf(f.Name, "`_` is not a valid rule group function name")
	}
	if f.Body == nil {
		return p.errorf(f, "unexpected empty function body")
	}
	if f.Type.Results != nil {
		return p.errorf(f.Type.Results, "rule group function should not return anything")
	}
	params := f.Type.Params.List
	if len(params) != 1 || len(params[0].Names) != 1 {
		return p.errorf(f.Type.Params, "rule group function should accept exactly 1 Matcher param")
	}
	// TODO(quasilyte): do an actual matcher param type check?
	matcher := params[0].Names[0].Name

	p.group = f.Name.Name

	p.itab.EnterScope()
	defer p.itab.LeaveScope()

	for _, stmt := range f.Body.List {
		if _, ok := stmt.(*ast.DeclStmt); ok {
			continue
		}
		stmtExpr, ok := stmt.(*ast.ExprStmt)
		if !ok {
			return p.errorf(stmt, "expected a %s method call, found %s", matcher, sprintNode(p.fset, stmt))
		}
		call, ok := stmtExpr.X.(*ast.CallExpr)
		if !ok {
			return p.errorf(stmt, "expected a %s method call, found %s", matcher, sprintNode(p.fset, stmt))
		}
		if err := p.parseCall(matcher, call); err != nil {
			return err
		}

	}

	return nil
}

func (p *rulesParser) parseCall(matcher string, call *ast.CallExpr) error {
	f := call.Fun.(*ast.SelectorExpr)
	x, ok := f.X.(*ast.Ident)
	if ok && x.Name == matcher {
		return p.parseStmt(f.Sel, call.Args)
	}

	return p.parseRule(matcher, call)
}

func (p *rulesParser) parseStmt(fn *ast.Ident, args []ast.Expr) error {
	switch fn.Name {
	case "Import":
		pkgPath, ok := p.toStringValue(args[0])
		if !ok {
			return p.errorf(args[0], "expected a string literal argument")
		}
		pkgName := path.Base(pkgPath)
		p.itab.Load(pkgName, pkgPath)
		return nil
	default:
		return p.errorf(fn, "unexpected %s method", fn.Name)
	}
}

func (p *rulesParser) parseRule(matcher string, call *ast.CallExpr) error {
	origCall := call
	var (
		matchArgs   *[]ast.Expr
		whereArgs   *[]ast.Expr
		suggestArgs *[]ast.Expr
		reportArgs  *[]ast.Expr
		atArgs      *[]ast.Expr
	)
	for {
		chain, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			break
		}
		switch chain.Sel.Name {
		case "Match":
			if matchArgs != nil {
				return p.errorf(chain.Sel, "Match() can't be repeated")
			}
			matchArgs = &call.Args
		case "Where":
			if whereArgs != nil {
				return p.errorf(chain.Sel, "Where() can't be repeated")
			}
			whereArgs = &call.Args
		case "Suggest":
			if suggestArgs != nil {
				return p.errorf(chain.Sel, "Suggest() can't be repeated")
			}
			suggestArgs = &call.Args
		case "Report":
			if reportArgs != nil {
				return p.errorf(chain.Sel, "Report() can't be repeated")
			}
			reportArgs = &call.Args
		case "At":
			if atArgs != nil {
				return p.errorf(chain.Sel, "At() can't be repeated")
			}
			atArgs = &call.Args
		default:
			return p.errorf(chain.Sel, "unexpected %s method", chain.Sel.Name)
		}
		call, ok = chain.X.(*ast.CallExpr)
		if !ok {
			break
		}
	}

	dst := p.res.universal
	proto := goRule{
		filename: p.filename,
		line:     p.fset.Position(origCall.Pos()).Line,
		group:    p.group,
	}
	var alternatives []string

	if matchArgs == nil {
		return p.errorf(origCall, "missing Match() call")
	}
	for _, arg := range *matchArgs {
		alternatives = append(alternatives, p.parseStringArg(arg))
	}

	if whereArgs != nil {
		proto.filter = p.parseFilter((*whereArgs)[0])
	}

	if suggestArgs != nil {
		proto.suggestion = p.parseStringArg((*suggestArgs)[0])
	}

	if reportArgs == nil {
		if suggestArgs == nil {
			return p.errorf(origCall, "missing Report() or Suggest() call")
		}
		proto.msg = "suggestion: " + proto.suggestion
	} else {
		proto.msg = p.parseStringArg((*reportArgs)[0])
	}

	if atArgs != nil {
		index, ok := (*atArgs)[0].(*ast.IndexExpr)
		if !ok {
			return p.errorf((*atArgs)[0], "expected %s[`varname`] expression", matcher)
		}
		arg, ok := p.toStringValue(index.Index)
		if !ok {
			return p.errorf(index.Index, "expected a string literal index")
		}
		proto.location = arg
	}

	for i, alt := range alternatives {
		rule := proto
		pat, err := gogrep.Parse(p.fset, alt)
		if err != nil {
			return p.errorf((*matchArgs)[i], "parse match pattern: %v", err)
		}
		rule.pat = pat
		cat := categorizeNode(pat.Expr)
		if cat == nodeUnknown {
			dst.uncategorized = append(dst.uncategorized, rule)
		} else {
			dst.categorizedNum++
			dst.rulesByCategory[cat] = append(dst.rulesByCategory[cat], rule)
		}
	}

	return nil
}

func (p *rulesParser) parseFilter(root ast.Expr) matchFilter {
	return p.parseFilterExpr(root)
}

func (p *rulesParser) errorf(n ast.Node, format string, args ...interface{}) parseError {
	loc := p.fset.Position(n.Pos())
	message := fmt.Sprintf("%s:%d: %s", loc.Filename, loc.Line, fmt.Sprintf(format, args...))
	return parseError(message)
}

func (p *rulesParser) parseStringArg(e ast.Expr) string {
	s, ok := p.toStringValue(e)
	if !ok {
		panic(p.errorf(e, "expected a string literal argument"))
	}
	return s
}

func (p *rulesParser) parseRegexpArg(e ast.Expr) *regexp.Regexp {
	patternString, ok := p.toStringValue(e)
	if !ok {
		panic(p.errorf(e, "expected a regexp pattern argument"))
	}
	re, err := regexp.Compile(patternString)
	if err != nil {
		panic(p.errorf(e, err.Error()))
	}
	return re
}

func (p *rulesParser) parseTypeStringArg(e ast.Expr) types.Type {
	typeString, ok := p.toStringValue(e)
	if !ok {
		panic(p.errorf(e, "expected a type string argument"))
	}
	typ, err := typeFromString(typeString)
	if err != nil {
		panic(p.errorf(e, "parse type expr: %v", err))
	}
	if typ == nil {
		panic(p.errorf(e, "can't convert %s into a type constraint yet", typeString))
	}
	return typ
}

func (p *rulesParser) parseFilterExpr(e ast.Expr) matchFilter {
	result := matchFilter{src: sprintNode(p.fset, e)}

	switch e := e.(type) {
	case *ast.ParenExpr:
		return p.parseFilterExpr(e.X)

	case *ast.UnaryExpr:
		x := p.parseFilterExpr(e.X)
		if e.Op == token.NOT {
			result.fn = makeNotFilter(result.src, x)
			return result
		}
		panic(p.errorf(e, "unsupported unary op: %s", result.src))

	case *ast.BinaryExpr:
		switch e.Op {
		case token.LAND:
			result.fn = makeAndFilter(p.parseFilterExpr(e.X), p.parseFilterExpr(e.Y))
			return result
		case token.LOR:
			result.fn = makeOrFilter(p.parseFilterExpr(e.X), p.parseFilterExpr(e.Y))
			return result
		case token.GEQ, token.LEQ, token.LSS, token.GTR, token.EQL, token.NEQ:
			operand := p.toFilterOperand(e.X)
			rhs := p.toFilterOperand(e.Y)
			rhsValue := p.types.Types[e.Y].Value
			if operand.path == "Type.Size" && rhsValue != nil {
				result.fn = makeTypeSizeConstFilter(result.src, operand.varName, e.Op, rhsValue)
				return result
			}
			if operand.path == "Value.Int" && rhsValue != nil {
				result.fn = makeValueIntConstFilter(result.src, operand.varName, e.Op, rhsValue)
				return result
			}
			if operand.path == "Value.Int" && rhs.path == "Value.Int" && rhs.varName != "" {
				result.fn = makeValueIntFilter(result.src, operand.varName, e.Op, rhs.varName)
				return result
			}
			if operand.path == "Text" && rhsValue != nil {
				result.fn = makeTextConstFilter(result.src, operand.varName, e.Op, rhsValue)
				return result
			}
			if operand.path == "Text" && rhs.path == "Text" && rhs.varName != "" {
				result.fn = makeTextFilter(result.src, operand.varName, e.Op, rhs.varName)
				return result
			}
		}
		panic(p.errorf(e, "unsupported binary op: %s", result.src))
	}

	operand := p.toFilterOperand(e)
	args := operand.args
	switch operand.path {
	default:
		panic(p.errorf(e, "unsupported expr: %s", result.src))

	case "File.Imports":
		pkgPath := p.parseStringArg(args[0])
		result.fn = makeFileImportsFilter(result.src, pkgPath)

	case "File.PkgPath.Matches":
		re := p.parseRegexpArg(args[0])
		result.fn = makeFilePkgPathMatchesFilter(result.src, re)

	case "File.Name.Matches":
		re := p.parseRegexpArg(args[0])
		result.fn = makeFileNameMatchesFilter(result.src, re)

	case "Pure":
		result.fn = makePureFilter(result.src, operand.varName)

	case "Const":
		result.fn = makeConstFilter(result.src, operand.varName)

	case "Addressable":
		result.fn = makeAddressableFilter(result.src, operand.varName)

	case "Type.Is", "Type.Underlying.Is":
		typeString, ok := p.toStringValue(args[0])
		if !ok {
			panic(p.errorf(args[0], "expected a string literal argument"))
		}
		ctx := typematch.Context{Itab: p.itab}
		pat, err := typematch.Parse(&ctx, typeString)
		if err != nil {
			panic(p.errorf(args[0], "parse type expr: %v", err))
		}
		underlying := operand.path == "Type.Underlying.Is"
		result.fn = makeTypeIsFilter(result.src, operand.varName, underlying, pat)

	case "Type.ConvertibleTo":
		dstType := p.parseTypeStringArg(args[0])
		result.fn = makeTypeConvertibleToFilter(result.src, operand.varName, dstType)

	case "Type.AssignableTo":
		dstType := p.parseTypeStringArg(args[0])
		result.fn = makeTypeAssignableToFilter(result.src, operand.varName, dstType)

	case "Type.Implements", "Type.Pointer.Implements":
		typeString, ok := p.toStringValue(args[0])
		if !ok {
			panic(p.errorf(args[0], "expected a string literal argument"))
		}
		n, err := parser.ParseExpr(typeString)
		if err != nil {
			panic(p.errorf(args[0], "parse type expr: %v", err))
		}
		var iface *types.Interface
		switch n := n.(type) {
		case *ast.Ident:
			if n.Name != `error` {
				panic(p.errorf(n, "only `error` unqualified type is recognized"))
			}
			iface = types.Universe.Lookup("error").Type().Underlying().(*types.Interface)
		case *ast.SelectorExpr:
			pkgName, ok := n.X.(*ast.Ident)
			if !ok {
				panic(p.errorf(n.X, "invalid package name"))
			}
			pkgPath, ok := p.itab.Lookup(pkgName.Name)
			if !ok {
				panic(p.errorf(n.X, "package %s is not imported", pkgName.Name))
			}
			pkg, err := p.stdImporter.Import(pkgPath)
			if err != nil {
				pkg, err = p.srcImporter.Import(pkgPath)
				if err != nil {
					panic(p.errorf(n, "can't load %s: %v", pkgPath, err))
				}
			}
			obj := pkg.Scope().Lookup(n.Sel.Name)
			if obj == nil {
				panic(p.errorf(n, "%s is not found in %s", n.Sel.Name, pkgPath))
			}
			iface, ok = obj.Type().Underlying().(*types.Interface)
			if !ok {
				panic(p.errorf(n, "%s is not an interface type", n.Sel.Name))
			}
		default:
			panic(p.errorf(args[0], "only qualified names (and `error`) are supported"))
		}
		toPointer := operand.path == "Type.Pointer.Implements"
		result.fn = makeTypeImplementsFilter(result.src, operand.varName, iface, toPointer)
	case "Text.Matches":
		re := p.parseRegexpArg(args[0])
		result.fn = makeTextMatchesFilter(result.src, operand.varName, re)

	case "Node.Is":
		typeString, ok := p.toStringValue(args[0])
		if !ok {
			panic(p.errorf(args[0], "expected a string literal argument"))
		}
		cat := categorizeNodeString(typeString)
		if cat == nodeUnknown {
			panic(p.errorf(args[0], "%s is not a valid go/ast type name", typeString))
		}
		result.fn = makeNodeIsFilter(result.src, operand.varName, cat)
	}

	if result.fn == nil {
		panic("bug: nil func for the filter") // Should never happen
	}
	return result
}

func (p *rulesParser) toStringValue(x ast.Node) (string, bool) {
	switch x := x.(type) {
	case *ast.BasicLit:
		if x.Kind != token.STRING {
			return "", false
		}
		s, err := strconv.Unquote(x.Value)
		if err != nil {
			return "", false
		}
		return s, true
	case ast.Expr:
		typ, ok := p.types.Types[x]
		if !ok || typ.Type.String() != "string" {
			return "", false
		}
		str := typ.Value.ExactString()
		str = str[1 : len(str)-1] // remove quotes
		return str, true
	}
	return "", false
}

func (p *rulesParser) toFilterOperand(e ast.Expr) filterOperand {
	var o filterOperand

	if call, ok := e.(*ast.CallExpr); ok {
		o.args = call.Args
		e = call.Fun
	}
	var path string
	for {
		if call, ok := e.(*ast.CallExpr); ok {
			e = call.Fun
			continue
		}
		selector, ok := e.(*ast.SelectorExpr)
		if !ok {
			break
		}
		if path == "" {
			path = selector.Sel.Name
		} else {
			path = selector.Sel.Name + "." + path
		}
		e = selector.X
	}

	o.path = path

	indexing, ok := e.(*ast.IndexExpr)
	if !ok {
		return o
	}
	mapIdent, ok := indexing.X.(*ast.Ident)
	if !ok {
		return o
	}
	o.mapName = mapIdent.Name
	indexString, _ := p.toStringValue(indexing.Index)
	o.varName = indexString

	return o
}

type filterOperand struct {
	mapName string
	varName string
	path    string
	args    []ast.Expr
}
