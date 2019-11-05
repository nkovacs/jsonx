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

FAQ
---

### Why yet another json package?

Most of the world uses camelCase or snake_case field names in JSON.
Go defaults to PascalCase as a side effect of exported identifiers, and makes it difficult to use something else, forcing you to litter your code with struct tags, which make your code less readable and harder to maintain.
The Go developers' [suggestion](https://github.com/golang/go/issues/23027#issuecomment-363232619) is to fork off or use a tool to automatically litter your code with struct tags (which isn't really a solution). Hence this fork.

### What about performance?

The standard library encoding/json package uses an internal cache of types, so it only needs to figure out how to encode a type once. This fork keeps that cache, but to fully benefit from it, you have to reuse the same `JSON` instance (or use the package-level functions, which are forwarded to a default instance). Methods like `OmitEmpty` create a new instance that shares the original's cache, so the performance impact of using those is minimal, but you can also reuse that instance to eliminate the overhead.

Performance is nearly identical to encoding/json, but slower than some of the fastest alternatives. See [benchmarks](#benchmarks).

Benchmarks
----------

Benchmarks taken from [jettison](https://github.com/wI2L/jettison).
Jettison encodes into a byte buffer to avoid allocating a byte slice,
so the jsonx benchmarks include one with Marshal and one with Encoder.Encode.

[BenchmarkSimplePayload](https://github.com/nkovacs/jsonbench/blob/master/bench_test.go#L45)

| package                                | time/op    | throughput   | bytes    | allocs      |
|----------------------------------------|------------|--------------|----------|-------------|
| encoding/json                          | 566ns ± 1% | 239MB/s ± 1% | 144 B/op | 1 allocs/op |
| jsonx                                  | 566ns ± 0% | 238MB/s ± 0% | 144 B/op | 1 allocs/op |
| jsonx encoder                          | 529ns ± 0% | 257MB/s ± 0% | 0 B/op   | 0 allocs/op |
| jsoniter                               | 598ns ± 1% | 226MB/s ± 1% | 152 B/op | 2 allocs/op |
| gojay                                  | 375ns ± 2% | 360MB/s ± 2% | 512 B/op | 1 allocs/op |
| jettison                               | 459ns ± 1% | 294MB/s ± 1% | 0 B/op   | 0 allocs/op |
| jettison NoUTF8Coercion NoHTMLEscaping | 397ns ± 0% | 340MB/s ± 0% | 0 B/op   | 0 allocs/op |

[BenchmarkComplexPayload](https://github.com/nkovacs/jsonbench/blob/master/bench_test.go#L148)

| package       | time/op     | throughput   | bytes    | allocs      |
|---------------|-------------|--------------|----------|-------------|
| encoding/json | 2170ns ± 1% | 178MB/s ± 1% | 416 B/op | 1 allocs/op |
| jsonx         | 2200ns ± 1% | 176MB/s ± 1% | 416 B/op | 1 allocs/op |
| jsonx encoder | 2120ns ± 1% | 183MB/s ± 1% | 0 B/op   | 0 allocs/op |
| jsoniter      | 2010ns ± 1% | 192MB/s ± 1% | 472 B/op | 3 allocs/op |
| jettison      | 1390ns ± 0% | 279MB/s ± 0% | 0 B/op   | 0 allocs/op |

[BenchmarkInterface](https://github.com/nkovacs/jsonbench/blob/master/bench_test.go#L269)

| package       | time/op     | throughput    | bytes  | allocs      |
|---------------|-------------|---------------|--------|-------------|
| encoding/json | 149ns ± 1%  | 53.5MB/s ± 2% | 8 B/op | 1 allocs/op |
| jsonx         | 152ns ± 0%  | 52.6MB/s ± 0% | 8 B/op | 1 allocs/op |
| jsonx encoder | 130ns ± 0%  | 69.0MB/s ± 1% | 0 B/op | 0 allocs/op |
| jsoniter      | 131ns ± 5%  | 61.0MB/s ± 5% | 8 B/op | 1 allocs/op |
| jettison      | 65.3ns ± 1% | 123MB/s ± 1%  | 0 B/op | 0 allocs/op |

[BenchmarkMap](https://github.com/nkovacs/jsonbench/blob/master/bench_test.go#L337)

| package         | time/op     | throughput    | bytes    | allocs       |
|-----------------|-------------|---------------|----------|--------------|
| encoding/json   | 1030ns ± 2% | 18.3MB/s ± 5% | 536 B/op | 13 allocs/op |
| jsonx           | 1030ns ± 2% | 18.4MB/s ± 2% | 536 B/op | 13 allocs/op |
| jsonx encoder   | 1020ns ± 1% | 19.6MB/s ± 1% | 504 B/op | 12 allocs/op |
| jsoniter        | 941ns ± 3%  | 20.2MB/s ± 3% | 680 B/op | 11 allocs/op |
| jettison sort   | 782ns ± 5%  | 24.3MB/s ± 5% | 496 B/op | 6 allocs/op  |
| jettison nosort | 328ns ± 2%  | 58.0MB/s ± 2% | 128 B/op | 2 allocs/op  |
