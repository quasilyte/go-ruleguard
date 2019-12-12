package ruleguard

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/quasilyte/go-ruleguard/internal/mvdan.cc/gogrep"
)

var filterPragmaRE = regexp.MustCompile(`^\$([^:]+):(.*)`)

type rulesParser struct {
	filename string
	fset     *token.FileSet
	sc       scanner
	res      *GoRuleSet
}

func (p *rulesParser) init(filename string, fset *token.FileSet, r io.Reader) error {
	// Maybe in the future we'll stop reading an entire rule file
	// eagerly, but it's easier this way for now.
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("%s: read error: %v", filename, err)
	}

	p.filename = filename
	p.fset = fset
	p.sc = scanner{lines: strings.Split(string(data), "\n")}
	p.res = &GoRuleSet{
		local:     &scopedGoRuleSet{},
		universal: &scopedGoRuleSet{},
	}

	return nil
}

func (p *rulesParser) parseTop() error {
	for {
		p.sc.scanLines(func(l string) bool { return l == "" })
		if !p.sc.canScan() {
			break
		}
		if err := p.parseRule(); err != nil {
			return err
		}
	}
	return nil
}

func (p *rulesParser) parseRule() error {
	dst := p.res.universal // Use "universal" set by default

	filters := map[string]submatchFilter{}
	proto := goRule{filters: filters}

	comment := p.sc.scanLines(func(l string) bool { return strings.HasPrefix(l, "//") })
	if len(comment) == 0 {
		return p.errorf(p.sc.i+1, "expected a comment")
	}
	for _, l := range comment {
		s := l.s[len("//"):]

		if m := filterPragmaRE.FindStringSubmatch(s); m != nil {
			name := m[1]
			expr := strings.TrimSpace(m[2])
			filter, err := p.makeFilter(expr)
			if err != nil {
				return p.errorf(l.num, "$%s: %v", name, err)
			}
			filters[name] = filter
			continue
		}

		switch {
		case strings.HasPrefix(s, "scope: local"):
			dst = p.res.local
		case strings.HasPrefix(s, "scope: universal"):
			dst = p.res.universal

		case strings.HasPrefix(s, "error:"):
			proto.severity = "error"
			proto.msg = strings.TrimSpace(s[len("error:"):])
		case strings.HasPrefix(s, "warning:"):
			proto.severity = "warning"
			proto.msg = strings.TrimSpace(s[len("warning:"):])
		case strings.HasPrefix(s, "information:"):
			proto.severity = "information"
			proto.msg = strings.TrimSpace(s[len("information:"):])
		case strings.HasPrefix(s, "hint:"):
			proto.severity = "hint"
			proto.msg = strings.TrimSpace(s[len("hint:"):])
		}
	}
	if proto.severity == "" {
		return nil
	}

	alternations := p.sc.scanLines(func(l string) bool {
		return !strings.HasPrefix(l, "//") && len(l) != 0
	})
	if len(alternations) == 0 {
		return p.errorf(p.sc.i+1, "expected one or more gogrep patterns")
	}
	for _, l := range alternations {
		rule := proto
		pat, err := gogrep.Parse(p.fset, l.s)
		if err != nil {
			return p.errorf(l.num, "gogrep parse: %v", err)
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

func (p *rulesParser) makeFilter(s string) (submatchFilter, error) {
	var filter submatchFilter
	expr, err := parser.ParseExpr(s)
	if err != nil {
		return filter, err
	}
	err = p.walkFilter(expr, &filter, false)
	return filter, err
}

func (p *rulesParser) walkFilter(expr ast.Expr, filter *submatchFilter, negate bool) error {
	badPredicate := func() error {
		return fmt.Errorf("unupported `%s` predicate", sprintNode(nil, expr))
	}

	switch expr := expr.(type) {
	case *ast.CallExpr:
		if filter.typePred != nil {
			// TODO(quasilyte): implement it.
			return fmt.Errorf("multi-type constraints are not implemented yet")
		}
		fn, ok := expr.Fun.(*ast.Ident)
		if !ok {
			return badPredicate()
		}
		switch fn.Name {
		case "is":
			if len(expr.Args) != 1 {
				return fmt.Errorf("is(type) expects exactly 1 param, %d given", len(expr.Args))
			}
			x := typeFromNode(expr.Args[0])
			wantIdentical := !negate
			filter.typePred = func(y types.Type) bool {
				return wantIdentical == types.Identical(x, y)
			}
		default:
			return badPredicate()
		}

	case *ast.Ident:
		if expr.Name == "pure" {
			if filter.pure != bool3unset {
				return fmt.Errorf("duplicated 'pure' constraint")
			}
			if negate {
				filter.pure = bool3false
			} else {
				filter.pure = bool3true
			}
		}
		if expr.Name == "constant" {
			if filter.constant != bool3unset {
				return fmt.Errorf("duplicated 'constant' constraint")
			}
			if negate {
				filter.constant = bool3false
			} else {
				filter.constant = bool3true
			}
		}
	case *ast.UnaryExpr:
		if expr.Op == token.NOT {
			return p.walkFilter(expr.X, filter, !negate)
		}
		return badPredicate()
	case *ast.BinaryExpr:
		if expr.Op == token.LAND {
			if err := p.walkFilter(expr.X, filter, negate); err != nil {
				return err
			}
			return p.walkFilter(expr.Y, filter, negate)
		}
		return badPredicate()
	default:
		return badPredicate()
	}

	return nil
}

func (p *rulesParser) errorf(lineNum int, format string, args ...interface{}) error {
	return &parseError{
		filename: p.filename,
		lineNum:  lineNum,
		msg:      fmt.Sprintf(format, args...),
	}
}

type parseError struct {
	filename string
	lineNum  int
	msg      string
}

func (e *parseError) Error() string {
	return fmt.Sprintf("%s:%d: %s", e.filename, e.lineNum, e.msg)
}
