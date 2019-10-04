// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsonx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

type Keys struct {
	Foo string
	Bar int
	Baz map[string]string
}

type OmitEmptyKeys struct {
	Foo string `json:",omitempty"`
	Bar int
}

func TestKeyEncodeFn(t *testing.T) {
	json := New(KeyEncodeFn(func(s string) string {
		r, z := utf8.DecodeRuneInString(s)
		return string(unicode.ToLower(r)) + s[z:] + s
	}))

	v := Keys{
		Foo: "foo",
		Bar: 42,
		Baz: map[string]string{
			"One":   "one",
			"two":   "two",
			"three": "Three",
		},
	}

	jsonV := []byte(`{"fooFoo":"foo","barBar":42,"bazBaz":{"One":"one","three":"Three","two":"two"}}`)

	t.Run("Marshal", func(t *testing.T) {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if !bytes.Equal(b, jsonV) {
			diff(t, b, jsonV)
		}
	})
	t.Run("Unmarshal", func(t *testing.T) {
		var v2 Keys
		err := json.Unmarshal(jsonV, &v2)
		if err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		if !reflect.DeepEqual(v2, v) {
			t.Errorf("mismatch\nhave: %#+v\nwant: %#+v", v2, v)
		}
	})
}

func TestJSONOmitEmpty(t *testing.T) {
	v := Keys{
		Foo: "foo",
		Bar: 0,
		Baz: map[string]string{},
	}

	jsonVEmpty := []byte(`{"Foo":"foo"}`)
	jsonV := []byte(`{"Foo":"foo","Bar":0,"Baz":{}}`)

	t.Run("true", func(t *testing.T) {
		t.Parallel()
		b, err := OmitEmpty().Marshal(v)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if !bytes.Equal(b, jsonVEmpty) {
			diff(t, b, jsonVEmpty)
		}
	})

	t.Run("false", func(t *testing.T) {
		t.Parallel()
		b, err := Marshal(v)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if !bytes.Equal(b, jsonV) {
			diff(t, b, jsonV)
		}
	})

	t.Run("with tag", func(t *testing.T) {
		v := OmitEmptyKeys{
			Foo: "",
			Bar: 0,
		}

		jsonVEmpty := []byte(`{}`)
		jsonV := []byte(`{"Bar":0}`)

		t.Run("true", func(t *testing.T) {
			t.Parallel()
			b, err := OmitEmpty().Marshal(v)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			if !bytes.Equal(b, jsonVEmpty) {
				diff(t, b, jsonVEmpty)
			}
		})

		t.Run("false", func(t *testing.T) {
			t.Parallel()
			b, err := Marshal(v)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			if !bytes.Equal(b, jsonV) {
				diff(t, b, jsonV)
			}
		})
	})
}

func TestJSONUseNumber(t *testing.T) {
	data := []byte(`2`)
	t.Run("true", func(t *testing.T) {
		t.Parallel()
		var v interface{}
		err := UseNumber().Unmarshal(data, &v)
		if err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		expected := json.Number("2")
		if !reflect.DeepEqual(expected, v) {
			t.Errorf("mismatch\nhave: %#+v\nwant: %#+v", v, expected)
		}
	})

	t.Run("true with decoder", func(t *testing.T) {
		t.Parallel()
		var v interface{}
		var buff bytes.Buffer
		decoder := UseNumber().NewDecoder(&buff)
		buff.Write(data)
		err := decoder.Decode(&v)
		if err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		expected := json.Number("2")
		if !reflect.DeepEqual(expected, v) {
			t.Errorf("mismatch\nhave: %#+v\nwant: %#+v", v, expected)
		}
	})

	t.Run("false", func(t *testing.T) {
		t.Parallel()
		var v interface{}
		err := Unmarshal(data, &v)
		if err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		expected := float64(2)
		if !reflect.DeepEqual(expected, v) {
			t.Errorf("mismatch\nhave: %#+v\nwant: %#+v", v, expected)
		}
	})

	t.Run("false with decoder", func(t *testing.T) {
		t.Parallel()
		var v interface{}
		var buff bytes.Buffer
		decoder := NewDecoder(&buff)
		buff.Write(data)
		err := decoder.Decode(&v)
		if err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		expected := float64(2)
		if !reflect.DeepEqual(expected, v) {
			t.Errorf("mismatch\nhave: %#+v\nwant: %#+v", v, expected)
		}
	})
}

func TestJSONDisallowUnknownFields(t *testing.T) {
	data := []byte(`{"x": 1}`)
	t.Run("true", func(t *testing.T) {
		t.Parallel()
		var v tx
		err := DisallowUnknownFields().Unmarshal(data, &v)
		expectedErr := fmt.Errorf("json: unknown field \"x\"")
		if err == nil || err.Error() != expectedErr.Error() {
			t.Fatalf("have: %v, want: %v", err, expectedErr)
		}
	})

	t.Run("true with decoder", func(t *testing.T) {
		t.Parallel()
		var v tx
		var buff bytes.Buffer
		decoder := DisallowUnknownFields().NewDecoder(&buff)
		buff.Write(data)
		err := decoder.Decode(&v)
		expectedErr := fmt.Errorf("json: unknown field \"x\"")
		if err == nil || err.Error() != expectedErr.Error() {
			t.Fatalf("have: %v, want: %v", err, expectedErr)
		}
	})

	t.Run("false", func(t *testing.T) {
		t.Parallel()
		var v tx
		err := Unmarshal(data, &v)
		if err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		expected := tx{}
		if !reflect.DeepEqual(expected, v) {
			t.Errorf("mismatch\nhave: %#+v\nwant: %#+v", v, expected)
		}
	})
}

func TestJSONEscapeHTML(t *testing.T) {
	data := `"<&>"`
	escaped := `"\"\u003c\u0026\u003e\""`
	unescaped := `"\"<&>\""`
	t.Run("true", func(t *testing.T) {
		t.Parallel()
		b, err := Marshal(data)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if string(b) != escaped {
			t.Fatalf("have: %v, want: %v", string(b), escaped)
		}
	})

	t.Run("true with encoder", func(t *testing.T) {
		t.Parallel()
		var buff bytes.Buffer
		encoder := NewEncoder(&buff)
		err := encoder.Encode(data)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		b := strings.TrimSpace(buff.String())
		if b != escaped {
			t.Fatalf("have: %v, want: %v", b, escaped)
		}
	})

	t.Run("false", func(t *testing.T) {
		t.Parallel()
		b, err := defaultJSON.EscapeHTML(false).Marshal(data)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if string(b) != unescaped {
			t.Fatalf("have: %v, want: %v", string(b), unescaped)
		}
	})

	t.Run("false with encoder", func(t *testing.T) {
		t.Parallel()
		var buff bytes.Buffer
		encoder := defaultJSON.EscapeHTML(false).NewEncoder(&buff)
		err := encoder.Encode(data)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		b := strings.TrimSpace(buff.String())
		if b != unescaped {
			t.Fatalf("have: %v, want: %v", b, unescaped)
		}
	})
}
