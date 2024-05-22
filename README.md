# go-json-enum

`go-json-enum` is a tool to automate the creation of methods that satisfy the json.Marshaler and json.Unmarshaler interfaces for common enum patterns. Given either a strng type or type that satisfies the fmt.Stringer interface T, `go-json-enum` will create a new self-contained Go source file implementing:

```go
func (t T) MarshalJSON() ([]byte, error)

func (t *T) UnmarshalJSON(data []byte) error
```
