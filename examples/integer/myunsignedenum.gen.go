// Code generated by "go-enum-codegen -type MyUnsignedEnum"; DO NOT EDIT.

package myenum

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)

// Scan implements sql.Scanner for MyUnsignedEnum
func (m *MyUnsignedEnum) Scan(value interface{}) error {
	u, ok := value.(uint)
	if !ok {
		return fmt.Errorf("failed to scan MyUnsignedEnum value: expected type `uint`, got `%T`", value)
	}
	switch u {
	case 1, 2, 3:
		*m = MyUnsignedEnum(u)
	default:
		*m = MyUnsignedEnumZero
	}

	return nil
}

// Value implements driver.Valuer for MyUnsignedEnum
func (m MyUnsignedEnum) Value() (driver.Value, error) {
	return uint(m), nil
}

// UnmarshalJSON implements json.Unmarshaler for MyUnsignedEnum
func (m *MyUnsignedEnum) UnmarshalJSON(data []byte) error {
	str := string(data)
	v, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to unmarshal MyUnsignedEnum value: could not convert `[]byte` to `uint`: %v", err)
	}
	u := uint(v)
	switch u {
	case 1, 2, 3:
		*m = MyUnsignedEnum(u)
	default:
		*m = MyUnsignedEnumZero
	}

	return nil
}

// MarshalJSON implements json.Marshaler for MyUnsignedEnum
func (m MyUnsignedEnum) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", uint(m))), nil
}