package goenumcodegen

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// TODO: more tests

func TestWriteSingleCaseStatement(t *testing.T) {
	receiver := "m"
	assignVar := "v"
	typeName := "MyEnum"
	tt := []struct {
		Name     string
		Input    []Value
		Kind     ValueType
		Expected string
	}{
		{
			Name: "string type, no default value",
			Input: []Value{
				{
					Name:   "MyEnumOne",
					StrVal: "one",
				},
				{
					Name:   "MyEnumTwo",
					StrVal: "two",
				},
				{
					Name:   "MyEnumThree",
					StrVal: "three",
				},
			},
			Kind:     TypeString,
			Expected: "\tcase \"one\", \"two\", \"three\":\n\t\t*m = MyEnum(v)\n",
		},
		{
			Name: "string type, default value",
			Input: []Value{
				{
					Name:   "MyEnumZero",
					StrVal: "",
				},
				{
					Name:   "MyEnumOne",
					StrVal: "one",
				},
				{
					Name:   "MyEnumTwo",
					StrVal: "two",
				},
				{
					Name:   "MyEnumThree",
					StrVal: "three",
				},
			},
			Kind:     TypeString,
			Expected: "\tcase \"one\", \"two\", \"three\":\n\t\t*m = MyEnum(v)\n",
		},
		{
			Name: "stringified numbers",
			Input: []Value{
				{
					Name:   "MyEnumZero",
					StrVal: "0",
				},
				{
					Name:   "MyEnumOne",
					StrVal: "1",
				},
				{
					Name:   "MyEnumTwo",
					StrVal: "2",
				},
				{
					Name:   "MyEnumThree",
					StrVal: "3",
				},
			},
			Kind:     TypeString,
			Expected: "\tcase \"0\", \"1\", \"2\", \"3\":\n\t\t*m = MyEnum(v)\n",
		},
		{
			Name: "number type, no default value",
			Input: []Value{
				{
					Name:   "MyEnumOne",
					StrVal: "1",
				},
				{
					Name:   "MyEnumTwo",
					StrVal: "2",
				},
				{
					Name:   "MyEnumThree",
					StrVal: "3",
				},
			},
			Kind:     TypeSigned,
			Expected: "\tcase 1, 2, 3:\n\t\t*m = MyEnum(v)\n",
		},
		{
			Name: "number type, default value",
			Input: []Value{
				{
					Name:   "MyEnumZero",
					StrVal: "0",
				},
				{
					Name:   "MyEnumOne",
					StrVal: "1",
				},
				{
					Name:   "MyEnumTwo",
					StrVal: "2",
				},
				{
					Name:   "MyEnumThree",
					StrVal: "3",
				},
			},
			Kind:     TypeSigned,
			Expected: "\tcase 1, 2, 3:\n\t\t*m = MyEnum(v)\n",
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			actual := WriteReadSingleCaseStatement(tc.Input, receiver, assignVar, typeName, tc.Kind)
			assert.Equal(t, tc.Expected, actual)
		})
	}
}

func TestWriteMultiCaseStatement(t *testing.T) {
	receiver := "m"
	tt := []struct {
		Name     string
		Input    []Value
		Expected string
	}{
		{
			Name: "no default value",
			Input: []Value{
				{
					Name:   "MyEnumOne",
					StrVal: "1",
				},
				{
					Name:   "MyEnumTwo",
					StrVal: "2",
				},
				{
					Name:   "MyEnumThree",
					StrVal: "3",
				},
			},
			Expected: "\tcase MyEnumOne.String():\n\t\t*m = MyEnumOne\n\tcase MyEnumTwo.String():\n\t\t*m = MyEnumTwo\n\tcase MyEnumThree.String():\n\t\t*m = MyEnumThree\n",
		},
		{
			Name: "default value",
			Input: []Value{
				{
					Name:   "MyEnumZero",
					StrVal: "0",
				},
				{
					Name:   "MyEnumOne",
					StrVal: "1",
				},
				{
					Name:   "MyEnumTwo",
					StrVal: "2",
				},
				{
					Name:   "MyEnumThree",
					StrVal: "3",
				},
			},
			Expected: "\tcase MyEnumOne.String():\n\t\t*m = MyEnumOne\n\tcase MyEnumTwo.String():\n\t\t*m = MyEnumTwo\n\tcase MyEnumThree.String():\n\t\t*m = MyEnumThree\n",
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			actual := WriteMultiCaseStatement(tc.Input, receiver)
			assert.Equal(t, tc.Expected, actual)
		})
	}
}
