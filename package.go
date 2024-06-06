package goenumcodegen

import (
	"go/ast"
	"go/types"
)

type Package struct {
	name  string
	defs  map[*ast.Ident]types.Object
	files []*File
}
