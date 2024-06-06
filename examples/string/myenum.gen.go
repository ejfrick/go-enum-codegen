// Code generated by "go-enum-codegen -type MyEnum"; DO NOT EDIT.

package myenum

import (
	"database/sql/driver"
	"fmt"
)

// Scan implements sql.Scanner for MyEnum
func (m *MyEnum) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan MyEnum value: expected type `string`, got `%T`", value)
	}
	switch str {
	case "One", "Three", "Two":
		*m = MyEnum(str)
	default:
		*m = MyEnumEmpty
	}

	return nil
}

// Value implements driver.Valuer for MyEnum
func (m MyEnum) Value() (driver.Value, error) {
	return string(m), nil
}

// UnmarshalJSON implements json.Unmarshaler for MyEnum
func (m *MyEnum) UnmarshalJSON(data []byte) error {
	str := string(data)
	switch str {
	case "One", "Three", "Two":
		*m = MyEnum(str)
	default:
		*m = MyEnumEmpty
	}

	return nil
}

// MarshalJSON implements json.Marshaler for MyEnum
func (m MyEnum) MarshalJSON() ([]byte, error) {
	return []byte(m), nil
}
