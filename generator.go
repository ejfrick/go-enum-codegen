package goenumcodegen

import (
	"bytes"
	"cmp"
	"fmt"
	"github.com/ejfrick/cuts"
	"go/ast"
	"go/format"
	"golang.org/x/tools/go/packages"
	"log"
	"slices"
	"strings"
)

type Generator struct {
	buf bytes.Buffer
	pkg *Package

	kinds []ValueType

	// generator config
	doJson      bool
	doScanValue bool
	errOnUnk    bool
	useString   bool
	debug       bool

	// per-type info
	// reset after each run
	isStringer   bool
	hasUnset     bool
	defaultValue *Value
}

func NewGenerator(opts ...Opt) *Generator {
	g := &Generator{
		doScanValue: true,
		doJson:      true,
	}

	g.options(opts...)
	g.defaults()

	g.logf("post-options fields: %#v", g)

	return g
}

func (g *Generator) defaults() {
	if !g.doJson && !g.doScanValue {
		g.doJson = true
		g.doScanValue = true
	}
}

type Opt func(g *Generator)

func WithOnlyJsonMethods() Opt {
	return func(g *Generator) {
		g.doScanValue = false
	}
}

func WithOnlySQLMethods() Opt {
	return func(g *Generator) {
		g.doJson = false
	}
}

func WithErrorOnUnknown() Opt {
	return func(g *Generator) {
		g.errOnUnk = true
	}
}

func WithUseStringer() Opt {
	return func(g *Generator) {
		g.useString = true
	}
}

func WithDebug() Opt {
	return func(g *Generator) {
		g.debug = true
	}
}

func (g *Generator) logf(format string, args ...interface{}) {
	if g.debug {
		log.Printf(format, args...)
	}
}

func (g *Generator) options(opts ...Opt) {
	for _, opt := range opts {
		opt(g)
	}
}

func (g *Generator) ParsePackage(patterns []string, tags []string) error {
	cfg := &packages.Config{
		Mode:       packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		Tests:      false,
		BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(tags, " "))},
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return err
	}
	if len(pkgs) != 1 {
		return fmt.Errorf("expected one package matching patterns %s, got %d", strings.Join(patterns, " "), len(pkgs))
	}
	g.addPackage(pkgs[0])

	return nil
}

func (g *Generator) addPackage(pkg *packages.Package) {
	g.pkg = &Package{
		name:  pkg.Name,
		defs:  pkg.TypesInfo.Defs,
		files: make([]*File, len(pkg.Syntax)),
	}

	for i, file := range pkg.Syntax {
		g.pkg.files[i] = &File{
			file: file,
			pkg:  g.pkg,
		}
	}
}

func (g *Generator) Printf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(&g.buf, format, args...)
}

func (g *Generator) WritePreambleAndImports(args []string) {
	body := g.buf.String()

	var s strings.Builder
	_, _ = s.WriteString(fmt.Sprintf("// Code generated by \"go-enum-codegen %s\"; DO NOT EDIT.\n\n", strings.Join(args, " ")))
	_, _ = s.WriteString(fmt.Sprintf("package %s\n\n", g.pkg.name))
	anySigned := cuts.AnyWhere(g.kinds, func(val ValueType) bool {
		return val == TypeSigned
	})
	anyUnsigned := cuts.AnyWhere(g.kinds, func(val ValueType) bool {
		return val == TypeUnsigned
	})
	if g.doScanValue || anySigned || anyUnsigned {
		_, _ = s.WriteString("import (\n")
		if g.doScanValue {
			_, _ = s.WriteString("\t\"database/sql/driver\"\n")
		}
		_, _ = s.WriteString("\t\"fmt\"\n")
		if (anySigned || anyUnsigned) && !g.useString && g.doJson {
			_, _ = s.WriteString("\t\"strconv\"\n")
		}
		_, _ = s.WriteString(")\n\n")
	} else {
		_, _ = s.WriteString("import \"fmt\"\n\n")
	}
	g.buf.Reset()
	g.Printf("%s%s", s.String(), body)
}

func (g *Generator) reset() {
	g.hasUnset = false
	g.isStringer = false
	g.defaultValue = nil
}

func (g *Generator) Format() ([]byte, error) {
	g.logf("Unformatted code:\n%s", g.buf.String())
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		return nil, err
	}
	return src, nil
}

func (g *Generator) Generate(typeName string) error {
	g.reset()
	g.logf("reset values for Generate run for type %s", typeName)
	values := make([]Value, 0, 100)
	for _, file := range g.pkg.files {
		file.typeName = typeName
		file.values = nil
		file.isStringer = false
		file.hasUnset = false
		if file.file != nil {
			g.logf("inspecting file %s", file.file.Name)
			ast.Inspect(file.file, file.GenDecl)
			values = append(values, file.values...)
		}
		if file.hasUnset {
			g.hasUnset = true
		}
	}

	if len(values) == 0 {
		return fmt.Errorf("no values defined for type %s", typeName)
	}
	g.logf("detected %d values", len(values))

	values = cuts.DedupeFunc(values, func(v Value) string {
		return v.StrVal
	})

	slices.SortStableFunc(values, func(a, b Value) int {
		return cmp.Compare(a.StrVal, b.StrVal)
	})

	kind := values[0].ValType
	recv := values[0].RecvName
	if recv == "" {
		recv = strings.ToLower(string(typeName[0]))
	}
	g.isStringer = values[0].IsStringer
	g.logf("data for type %s: kind: %s, receiver: %s, isStringer: %t", typeName, kind, recv, g.isStringer)

	if (kind == TypeSigned || kind == TypeUnsigned) && g.useString && !g.isStringer {
		return fmt.Errorf("type %s does not implement fmt.Stringer", typeName)
	}

	g.kinds = append(g.kinds, kind)

	var defaultValue Value
	if kind == TypeString {
		defaultValue = Value{StrVal: "\"\""}
	} else {
		defaultValue = Value{StrVal: "0"}
	}

	index, exists := slices.BinarySearchFunc(values, defaultValue, func(a, b Value) int {
		return cmp.Compare(a.StrVal, b.StrVal)
	})

	if exists {
		v := values[index]
		g.logf("detected default value %#v", v)
		g.defaultValue = &v
		if !g.errOnUnk {
			g.logf("removing default value from value list")
			values = slices.Delete(values, index, index+1)
		}
	}

	if g.doScanValue {
		g.logf("starting sql.Scanner & driver.Valuer run")
		g.writeScannerValuer(recv, values, kind, typeName)
	}

	if g.doJson {
		g.logf("starting json.Marshaler and json.Unmarshaler run")
		g.writeMarshalerUnmarshaler(recv, values, kind, typeName)
	}

	return nil
}

func (g *Generator) writeScannerValuer(recv string, values []Value, kind ValueType, typeName string) {
	assgnVar, convType := g.getReadAssignVarAndConvType(kind)
	g.logf("using assignment variable %s and will convert to type %s for Scan method", assgnVar, convType)
	g.Printf("// Scan implements sql.Scanner for %s\n", typeName)
	g.Printf("func (%s *%s) Scan(value interface{}) error {\n", recv, typeName)
	g.writeScannerTypeAssertionStmnt("scan", assgnVar, convType, typeName)
	g.logf("wrote type assertion statement")
	g.writeReadCaseStatement(recv, values, kind, assgnVar, typeName)
	g.logf("wrote case statement")
	g.writeReadDefaultCase("scan", recv, assgnVar, typeName)
	g.logf("wrote default case statement")
	g.writeReadCloser()
	g.Printf("// Value implements driver.Valuer for %s\n", typeName)
	g.writeValuerBody(recv, kind, typeName)
	g.logf("wrote Value method")
}

func (g *Generator) writeMarshalerUnmarshaler(recv string, values []Value, kind ValueType, typeName string) {
	assgnVar, convType := g.getReadAssignVarAndConvType(kind)
	g.logf("using assignment variable %s and will convert to type %s for UnmarshalJSON method", assgnVar, convType)
	g.Printf("// UnmarshalJSON implements json.Unmarshaler for %s\n", typeName)
	g.Printf("func (%s *%s) UnmarshalJSON(data []byte) error {\n", recv, typeName)
	g.writeUnmarshalerTypeConversionStmnt(assgnVar, convType, "unmarshal", typeName)
	g.logf("wrote type conversion statement")
	g.writeReadCaseStatement(recv, values, kind, assgnVar, typeName)
	g.logf("wrote case statement")
	g.writeReadDefaultCase("unmarshal", recv, assgnVar, typeName)
	g.logf("wrote default case statement")
	g.writeReadCloser()
	g.Printf("// MarshalJSON implements json.Marshaler for %s\n", typeName)
	g.writeMarshalerBody(recv, convType, typeName)
	g.logf("wrote MarshalJSON method")
}

func (g *Generator) writeMarshalerBody(recv string, convType string, typeName string) {
	g.Printf("func (%s %s) MarshalJSON() ([]byte, error) {\n", recv, typeName)
	g.Printf("\treturn []byte(")
	switch {
	case g.useString && g.isStringer:
		g.Printf("%s.String()", recv)
		g.logf("returning %s.String(), nil for MarshalJSON", typeName)
	case convType == "string":
		g.Printf("%s", recv)
		g.logf("returning %s, nil for MarshalJSON", typeName)
	default:
		g.Printf("fmt.Sprintf(\"%%d\", %s(%s))", convType, recv)
		g.logf("returning fmt.Sprintf'd %s, nil for MarshalJSON", typeName)
	}
	g.Printf("), nil\n")
	g.Printf("}\n\n")
}

func (g *Generator) writeValuerBody(recv string, kind ValueType, typeName string) {
	g.Printf("func (%s %s) Value() (driver.Value, error) {\n", recv, typeName)
	var returnStmt strings.Builder
	switch {
	case g.useString && g.isStringer:
		_, _ = returnStmt.WriteString(recv)
		_, _ = returnStmt.WriteString(".String()")
		g.logf("Valuer will return %s.String()", typeName)
	case kind == TypeString:
		_, _ = returnStmt.WriteString("string(")
		_, _ = returnStmt.WriteString(recv)
		_, _ = returnStmt.WriteString(")")
		g.logf("Valuer will return string(%s)", typeName)
	case kind == TypeSigned:
		_, _ = returnStmt.WriteString("int(")
		_, _ = returnStmt.WriteString(recv)
		_, _ = returnStmt.WriteString(")")
		g.logf("Valuer will return int(%s)", typeName)
	default:
		_, _ = returnStmt.WriteString("uint(")
		_, _ = returnStmt.WriteString(recv)
		_, _ = returnStmt.WriteString(")")
		g.logf("Valuer will return uint(%s)", typeName)
	}
	g.Printf("\treturn %s, nil\n", returnStmt.String())
	g.Printf("}\n\n")
}

func (g *Generator) writeReadCloser() {
	g.Printf("\t}\n")
	g.Printf("\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n\n")
}

func (g *Generator) writeReadDefaultCase(method string, recv string, assgnVar string, typeName string) {
	g.Printf("\tdefault:\n")
	switch {
	case !g.errOnUnk && g.defaultValue != nil && !g.hasUnset:
		g.logf("writing default statement to assign to default value")
		g.Printf("\t\t*%s = %s\n", recv, g.defaultValue.Name)
	default:
		g.logf("writing default statement to return error")
		g.Printf("\t\treturn fmt.Errorf(\"failed to %s %s value: unrecognized value `%%v`\", %s)\n", method, typeName, assgnVar)
	}
}

func (g *Generator) writeReadCaseStatement(recv string, values []Value, kind ValueType, assgnVar string, typeName string) {
	var stmnt string
	switch {
	case (kind == TypeSigned || kind == TypeUnsigned) && g.useString && g.isStringer:
		g.logf("writing multi-case statement")
		stmnt = WriteMultiCaseStatement(values, recv)
	default:
		g.logf("writing single case statement")
		stmnt = WriteReadSingleCaseStatement(values, recv, assgnVar, typeName, kind)

	}
	g.Printf(stmnt)
}

func (g *Generator) writeScannerTypeAssertionStmnt(method string, assgnVar string, convType string, typeName string) {
	g.Printf("\t%s, ok := value.(%s)\n", assgnVar, convType)
	g.Printf("\tif !ok {\n")
	g.Printf("\t\treturn fmt.Errorf(\"failed to %s %s value: expected type `%s`, got `%%T`\", value)\n", method, typeName, convType)
	g.Printf("\t}\n")
	g.Printf("\tswitch %s {\n", assgnVar)
}

func (g *Generator) writeUnmarshalerTypeConversionStmnt(assgnVar string, convType string, method string, typeName string) {
	g.Printf("\tstr := string(data)\n")
	g.logf("converting []byte to string")
	if convType == "int" {
		g.Printf("\tv, err := strconv.ParseInt(str, 10, 64)\n")
		g.logf("using strconv.ParseInt")
	} else if convType == "uint" {
		g.Printf("\tv, err := strconv.ParseUint(str, 10, 64)\n")
		g.logf("using strconv.ParseUint")
	}
	if convType == "uint" || convType == "int" {
		g.Printf("\tif err != nil {\n")
		g.Printf("\t\treturn fmt.Errorf(\"failed to %s %s value: could not convert `[]byte` to `%s`: %%v\", err)\n", method, typeName, convType)
		g.Printf("\t}\n")
		g.Printf("\t%s := %s(v)\n", assgnVar, convType)
	}
	g.Printf("\tswitch %s {\n", assgnVar)
}

func (g *Generator) getReadAssignVarAndConvType(kind ValueType) (string, string) {
	var assgnVar string
	var t string
	switch {
	case kind == TypeString, g.useString && g.isStringer:
		t = "string"
		assgnVar = "str"
	case kind == TypeSigned:
		t = "int"
		assgnVar = "i"
	default:
		t = "uint"
		assgnVar = "u"
	}
	return assgnVar, t
}

func WriteReadSingleCaseStatement(values []Value, receiver string, assgnVar string, typeName string, kind ValueType) string {
	var s strings.Builder
	_, _ = s.WriteString("\tcase ")
	var valValues []string
	for _, value := range values {
		var val string
		if kind == TypeString {
			val = value.StrVal
		} else {
			val = strings.Trim(value.StrVal, `"`)
		}
		valValues = append(valValues, val)
	}
	vals := strings.Join(valValues, ", ")
	_, _ = s.WriteString(vals)
	_, _ = s.WriteString(":\n")
	_, _ = s.WriteString(fmt.Sprintf("\t\t*%s = %s(%s)\n", receiver, typeName, assgnVar))
	return s.String()
}

func WriteMultiCaseStatement(values []Value, receiver string) string {
	var s strings.Builder
	for _, value := range values {
		_, _ = s.WriteString(fmt.Sprintf("\tcase %s.String():\n", value.Name))
		_, _ = s.WriteString(fmt.Sprintf("\t\t*%s = %s\n", receiver, value.Name))
	}
	return s.String()
}
