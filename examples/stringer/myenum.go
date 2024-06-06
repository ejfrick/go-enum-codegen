package myenum

type MyEnum int

const (
	MyEnumZero MyEnum = iota
	MyEnumOne
	MyEnumTwo
	MyEnumThree
)

func (e MyEnum) String() string {
	var s string
	switch e {
	case MyEnumZero:
		s = "zero"
	case MyEnumOne:
		s = "one"
	case MyEnumTwo:
		s = "two"
	case MyEnumThree:
		s = "three"
	default:
		s = MyEnumZero.String()
	}

	return s
}
