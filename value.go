package goenumcodegen

type ValueType string

const (
	TypeString   ValueType = "string"
	TypeSigned   ValueType = "int"
	TypeUnsigned ValueType = "uint"
)

type Value struct {
	Name       string
	ValType    ValueType
	StrVal     string
	IsStringer bool
	RecvName   string
}
