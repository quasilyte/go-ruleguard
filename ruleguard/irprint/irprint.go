package irprint

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"reflect"

	"github.com/quasilyte/go-ruleguard/ruleguard/ir"
)

func File(w io.Writer, f *ir.File) {
	p := printer{}
	p.printFile(f)
	pretty, err := format.Source(p.buf.Bytes())
	if err != nil {
		fmt.Println(p.buf.String())
		panic(err)
	}
	_, _ = w.Write(pretty)
}

type printer struct {
	buf bytes.Buffer
}

func (p *printer) writef(format string, args ...interface{}) {
	fmt.Fprintf(&p.buf, format, args...)
}

func (p *printer) printFile(f *ir.File) {
	p.writef("ir.File{\n")

	p.writef("PkgPath: %q,\n", f.PkgPath)

	p.writef("CustomDecls: []string{\n")
	for _, src := range f.CustomDecls {
		p.writef("%q,\n", src)
	}
	p.writef("},\n")

	p.writef("BundleImports: []ir.BundleImport{\n")
	for _, imp := range f.BundleImports {
		p.writef("Line: %d,\n", imp.Line)
		p.writef("PkgPath: %q,\n", imp.PkgPath)
		p.writef("Prefix: %q,\n", imp.PkgPath)
	}
	p.writef("},\n")

	p.printReflectElem("RuleGroups", reflect.ValueOf(f.RuleGroups), false)

	p.writef("}\n")
}

func (p *printer) printReflectElem(key string, v reflect.Value, insideList bool) {
	if p.printReflectElemNoNewline(key, v, insideList) {
		p.buf.WriteByte('\n')
	}
}

func (p *printer) printReflectElemNoNewline(key string, v reflect.Value, insideList bool) bool {
	if v.IsZero() {
		return false
	}

	if key != "" {
		p.writef("%s: ", key)
	}

	if v.Type().Name() == "FilterOp" {
		p.writef("ir.Filter%sOp,", v.Interface().(ir.FilterOp).String())
		return true
	}

	// There are tons of these, print them in a compact way.
	if v.Type().Name() == "PatternString" {
		v := v.Interface().(ir.PatternString)
		if !insideList {
			p.buf.WriteString("ir.PatternString")
		}
		p.writef("{Line: %d, Value: %#v},", v.Line, v.Value)
		return true
	}

	if v.Type().Name() == "FilterExpr" {
		v := v.Interface().(ir.FilterExpr)
		if v.Op == ir.FilterStringOp || v.Op == ir.FilterVarPureOp || v.Op == ir.FilterVarTextOp {
			if !insideList {
				p.buf.WriteString("ir.FilterExpr")
			}
			p.writef("{Line: %d, Op: ir.Filter%sOp, Src: %#v, Value: %#v},",
				v.Line, v.Op.String(), v.Src, v.Value.(string))
			return true
		}
	}

	switch v.Type().Kind() {
	case reflect.Struct:
		if !insideList {
			p.buf.WriteString(v.Type().String())
		}
		p.writef("{\n")
		for i := 0; i < v.NumField(); i++ {
			p.printReflectElem(v.Type().Field(i).Name, v.Field(i), false)
		}
		p.writef("},")

	case reflect.Slice:
		if isCompactSlice(v) {
			p.writef("%s{", v.Type())
			for j := 0; j < v.Len(); j++ {
				p.printReflectElemNoNewline("", v.Index(j), true)
			}
			p.writef("},")
		} else {
			p.writef("%s{\n", v.Type().String())
			for j := 0; j < v.Len(); j++ {
				p.printReflectElem("", v.Index(j), true)
			}
			p.writef("},")
		}

	default:
		switch val := v.Interface().(type) {
		case int64:
			p.writef("int64(%v),", val)
		default:
			p.writef("%#v,", val)
		}
	}

	return true
}

func isCompactSlice(v reflect.Value) bool {
	switch v.Type().Elem().Kind() {
	case reflect.String, reflect.Int:
		return v.Len() <= 4
	default:
		return v.Len() <= 1
	}
}
