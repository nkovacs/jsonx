jsonx
=====

jsonx is an improved fork of the go standard library json package.

Features
--------

### Configurable field naming convention

Instead of struct tags, jsonx lets you specify a function that is applied to all field names automatically:

```go
json := jsonx.New(jsonx.KeyEncodeFn(func(s string) string {
	r, z := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[z:]
}))

b, _ := json.Marshal(struct{
	FirstName string
	LastName  string
	Email     string
}{
	FirstName: "John",
	LastName:  "Doe",
	Email:     "jdoe@example.com",
})
fmt.Println(string(b))
// {"firstName":"John","lastName":"Doe","email":"jdoe@example.com"}
```

### OmitEmpty

Instead of using the `omitempty` struct tag on all your fields, you can configure the json encoder to omit empty fields globally or for a single `Marshal` call:

```go
user := struct {
	FirstName string
	LastName  string
	Email     string
	Nickname  string
}{
	FirstName: "John",
	LastName:  "Doe",
	Email:     "jdoe@example.com",
}
b, _ := jsonx.OmitEmpty().Marshal(user)
fmt.Println(string(b))
// {"FirstName":"John","LastName":"Doe","Email":"jdoe@example.com"}

// alternatively, with a JSON instance
json := jsonx.New(jsonx.KeyEncodeFn(func(s string) string {
	r, z := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[z:]
}))
b, _ = json.OmitEmpty().Marshal(user)
fmt.Println(string(b))
// {"firstName":"John","lastName":"Doe","email":"jdoe@example.com"}
```

### Better API

If you want to unmarshal numbers as json.Number instead of float64 or if you want to get an error in case the json input contains a field that is not present in the destination struct, you have to create a `json.Decoder` and set its options. Aside from forcing you to use a different API using `io.Reader`, [json.Decoder is designed for JSON streams, not single JSON objects](https://ahmet.im/blog/golang-json-decoder-pitfalls/).

jsonx allows you to set these options for Marshal and Unmarshal:

```go
var v interface{}
jsonx.UseNumber().Unmarshal([]byte("2"), &v)
fmt.Printf("%T %[1]v\n", v)
// json.Number 2

var t struct{
	Foo int
}
err := jsonx.DisallowUnknownFields().Unmarshal([]byte(`{"foo": 1, "bar": 2}`), &t)
fmt.Println(err)
// json: unknown field "bar"
```

### Compatibility

jsonx is not meant as a full replacement to encoding/json. It reuses as much of encoding/json as it can, including types such as `json.Number` and `json.RawMessage`, and does not duplicate `json.Compact`, `json.Indent`, `json.Valid` and `json.HTMLEscape`.

All errors are the same as encoding/json except `json.SyntaxError`, which has an unexported field. jsonx uses `jsonx.SyntaxError` instead.

jsonx respects json struct tags, which can override both the key encoding function and OmitEmpty.

jsonx also respects `UnmarshalJSON`, `MarshalJSON`, `UnmarshalText` and `MarshalText`, but note that it cannot control what happens in those.
Passing down options to a type's own Marshal or Unmarshal method is a [complicated problem](https://github.com/golang/go/issues/14750#issuecomment-422238315), and one that this package does not try to solve.