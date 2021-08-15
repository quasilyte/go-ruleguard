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

	p.printReflectElem("RuleGroups", reflect.ValueOf(f.RuleGroups))

	p.writef("}\n")
}

func (p *printer) printReflectElem(key string, v reflect.Value) {
	if v.IsZero() {
		return
	}

	if key != "" {
		p.writef("%s: ", key)
	}

	if v.Type().Name() == "FilterOp" {
		p.writef("ir.Filter%sOp,\n", v.Interface().(ir.FilterOp).String())
		return
	}

	if v.Type().Kind() == reflect.Struct {
		p.writef("%s{\n", v.Type().String())
		for i := 0; i < v.NumField(); i++ {
			p.printReflectElem(v.Type().Field(i).Name, v.Field(i))
		}
		p.writef("},\n")
	} else if v.Type().Kind() == reflect.Slice {
		p.writef("%s{\n", v.Type().String())
		for j := 0; j < v.Len(); j++ {
			p.printReflectElem("", v.Index(j))
		}
		p.writef("},\n")
	} else {
		switch val := v.Interface().(type) {
		case int64:
			p.writef("int64(%v),\n", val)
		default:
			p.writef("%#v,\n", val)
		}
	}
}
