package myenum

type MySignedEnum int

const (
	MySignedEnumZero MySignedEnum = iota
	MySignedEnumOne
	MySignedEnumTwo
	MySignedEnumThree
)

type MyUnsignedEnum uint

const (
	MyUnsignedEnumZero MyUnsignedEnum = iota
	MyUnsignedEnumOne
	MyUnsignedEnumTwo
	MyUnsignedEnumThree
)
