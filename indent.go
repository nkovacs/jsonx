// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsonx

import "bytes"

func compact(dst *bytes.Buffer, src []byte, escape bool) error {
	origLen := dst.Len()
	var scan scanner
	scan.reset()
	start := 0
	for i, c := range src {
		if escape && (c == '<' || c == '>' || c == '&') {
			if start < i {
				dst.Write(src[start:i])
			}
			dst.WriteString(`\u00`)
			dst.WriteByte(hex[c>>4])
			dst.WriteByte(hex[c&0xF])
			start = i + 1
		}
		// Convert U+2028 and U+2029 (E2 80 A8 and E2 80 A9).
		if escape && c == 0xE2 && i+2 < len(src) && src[i+1] == 0x80 && src[i+2]&^1 == 0xA8 {
			if start < i {
				dst.Write(src[start:i])
			}
			dst.WriteString(`\u202`)
			dst.WriteByte(hex[src[i+2]&0xF])
			start = i + 3
		}
		v := scan.step(&scan, c)
		if v >= scanSkipSpace {
			if v == scanError {
				break
			}
			if start < i {
				dst.Write(src[start:i])
			}
			start = i + 1
		}
	}
	if scan.eof() == scanError {
		dst.Truncate(origLen)
		return scan.err
	}
	if start < len(src) {
		dst.Write(src[start:])
	}
	return nil
}
