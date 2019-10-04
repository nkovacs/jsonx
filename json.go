// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsonx

import "sync"

// JSON is a json encoder/decoder.
// It is safe for concurrent use by multiple goroutines.
type JSON struct {
	// keyEncodeFn is applied to struct field names to create object keys.
	keyEncodeFn           func(string) string
	fieldCache            *sync.Map // map[reflect.Type]structFields
	encoderCache          *sync.Map // map[reflect.Type]encoderFunc
	omitEmpty             bool
	useNumber             bool
	disallowUnknownFields bool
	dontEscapeHTML        bool
}

var defaultJSON = &JSON{
	fieldCache:   &sync.Map{},
	encoderCache: &sync.Map{},
}

// Options are used to customize a JSON encoder/decoder.
type Options interface {
	// SetKeyEncodeFn sets the function that is applied to struct field names
	// to create object keys when marshaling.
	// It is also used to match incoming object keys to struct fields when unmarshaling,
	// by encoding the struct fields and then matching them case insensitively.
	SetKeyEncodeFn(func(string) string)
}

// Option is a JSON encoder/decoder option.
type Option func(Options)

type jsonOptionWrapper struct {
	json *JSON
}

func (w *jsonOptionWrapper) SetKeyEncodeFn(fn func(string) string) {
	w.json.keyEncodeFn = fn
}

// KeyEncodeFn sets the key encoding function
// when creating a new JSON encoder/decoder.
func KeyEncodeFn(fn func(string) string) Option {
	return func(opt Options) {
		opt.SetKeyEncodeFn(fn)
	}
}

// New creates a new JSON encoder/decoder.
//
// The encoder has an internal cache,
// so it should be reused for best performance.
// Changing the key encoding function is not possible
// because it would require invalidating the cache.
func New(opts ...Option) *JSON {
	json := &JSON{
		fieldCache:   &sync.Map{},
		encoderCache: &sync.Map{},
	}
	w := &jsonOptionWrapper{json: json}
	for _, opt := range opts {
		opt(w)
	}
	return json
}

// OmitEmpty specifies that fields with an empty value
// should be omitted from encoding.
// It returns a copy of the original JSON encoder/decoder, sharing its cache.
func (j *JSON) OmitEmpty() *JSON {
	j2 := *j
	j2.omitEmpty = true
	return &j2
}

// OmitEmpty specifies that fields with an empty value
// should be omitted from encoding.
// It returns a copy of the default JSON encoder/decoder, sharing its cache.
func OmitEmpty() *JSON {
	return defaultJSON.OmitEmpty()
}

// UseNumber causes the decoder to unmarshal a number into an interface{} as a
// json.Number instead of as a float64.
// It returns a copy of the original JSON encoder/decoder, sharing its cache.
func (j *JSON) UseNumber() *JSON {
	j2 := *j
	j2.useNumber = true
	return &j2
}

// UseNumber causes the decoder to unmarshal a number into an interface{} as a
// json.Number instead of as a float64.
// It returns a copy of the default JSON encoder/decoder, sharing its cache.
func UseNumber() *JSON {
	return defaultJSON.UseNumber()
}

// DisallowUnknownFields causes the decoder to return an error when the destination
// is a struct and the input contains object keys which do not match any
// non-ignored, exported fields in the destination.
// It returns a copy of the original JSON encoder/decoder, sharing its cache.
func (j *JSON) DisallowUnknownFields() *JSON {
	j2 := *j
	j2.disallowUnknownFields = true
	return &j2
}

// DisallowUnknownFields causes the decoder to return an error when the destination
// is a struct and the input contains object keys which do not match any
// non-ignored, exported fields in the destination.
// It returns a copy of the default JSON encoder/decoder, sharing its cache.
func DisallowUnknownFields() *JSON {
	return defaultJSON.DisallowUnknownFields()
}

// EscapeHTML specifies whether problematic HTML characters
// should be escaped inside JSON quoted strings.
// The default behavior is to escape &, <, and > to \u0026, \u003c, and \u003e
// to avoid certain safety problems that can arise when embedding JSON in HTML.
//
// In non-HTML settings where the escaping interferes with the readability
// of the output, EscapeHTML(false) disables this behavior.
// It returns a copy of the original JSON encoder/decoder, sharing its cache.
func (j *JSON) EscapeHTML(on bool) *JSON {
	j2 := *j
	j2.dontEscapeHTML = !on
	return &j2
}
