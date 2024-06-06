package goenumcodegen

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"log"
)

type File struct {
	pkg          *Package
	file         *ast.File
	typeName     string
	values       []Value
	isStringer   bool
	receiverName string
	noEmpty      bool
}

func (f *File) GenDecl(node ast.Node) bool {
	decl, ok := node.(*ast.GenDecl)
	if !ok || decl.Tok != token.CONST {
		return true
	}

	typ := ""
	for _, spec := range decl.Specs {
		vspec := spec.(*ast.ValueSpec) // Guaranteed to succeed as this is CONST.
		if vspec.Type == nil && len(vspec.Values) > 0 {
			typ = ""
			ce, ok := vspec.Values[0].(*ast.CallExpr)
			if !ok {
				continue
			}
			id, ok := ce.Fun.(*ast.Ident)
			if !ok {
				continue
			}
			typ = id.Name
		}
		if vspec.Type != nil {
			ident, ok := vspec.Type.(*ast.Ident)
			if !ok {
				continue
			}
			typ = ident.Name
		}
		if typ != f.typeName {
			// This is not the type we're looking for.
			continue
		}
		for _, name := range vspec.Names {
			if name.Name == "_" {
				continue
			}
			obj, ok := f.pkg.defs[name]
			if !ok {
				log.Fatalf("no value for constant %s", name)
			}
			info := obj.Type().Underlying().(*types.Basic).Info()
			if info&types.IsInteger == 0 && info&types.IsString == 0 {
				log.Fatalf("can't handle non-interger, non-string constant type %s", typ)
			}
			value := obj.(*types.Const).Val()
			v := Value{
				Name:   name.Name,
				StrVal: value.String(),
			}
			if value.Kind() == constant.String {
				v.ValType = TypeString
			} else {
				v.IsStringer = IsStringer(obj.Type().(*types.Named))
				if info&types.IsUnsigned != 0 {
					v.ValType = TypeUnsigned
				} else {
					v.ValType = TypeSigned
				}
			}
			v.RecvName = GetReceiver(obj.Type().(*types.Named))
			f.values = append(f.values, v)
		}
	}
	return false
}

func GetReceiver(obj *types.Named) string {
	for i := 0; i < obj.NumMethods(); i++ {
		method := obj.Method(i)
		if method.Type().(*types.Signature).Recv() != nil && method.Type().(*types.Signature).Recv().Name() != "" {
			return method.Type().(*types.Signature).Recv().Name()
		}
	}

	return ""
}

func IsStringer(obj *types.Named) bool {
	for i := 0; i < obj.NumMethods(); i++ {
		m := obj.Method(i)
		if m.Name() == "String" && m.Type().(*types.Signature).Results().Len() == 1 && m.Type().(*types.Signature).Results().At(0).Type().String() == "string" {
			return true
		}
	}

	return false
}
