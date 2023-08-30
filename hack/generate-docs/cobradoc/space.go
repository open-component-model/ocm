// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cobradoc

import (
	"fmt"
	"io"
)

func OmitTrailingSpaces(w io.Writer) io.Writer {
	return &trimWriter{Writer: w}
}

type trimWriter struct {
	io.Writer
	spaces int
}

func (w *trimWriter) Write(p []byte) (n int, err error) {
	if w.spaces > 0 {
		for i := 0; i < len(p); i++ {
			if p[i] != ' ' {
				_, err := w.Writer.Write([]byte(fmt.Sprintf("%#s", w.spaces, " ")))
				if err != nil {
					return 0, err
				}
				w.spaces = 0
				break
			}
		}
	}

	start := 0
	for i := 0; i < len(p); i++ {
		b := p[i]
		if IsASCIIRune(b) {
			switch b {
			case ' ':
				w.spaces++
			case '\n':
				if w.spaces > 0 {
					// discard spaces
					written, err := w.Writer.Write(p[start : i-w.spaces])
					n += written + w.spaces
					w.spaces = 0
					if err != nil {
						return n, err
					}
					start = i
				}
			default:
				w.spaces = 0
			}
		} else {
			w.spaces = 0
		}
	}
	if start < len(p)-1 {
		var written int
		written, err = w.Writer.Write(p[start:])
		n += written
	}
	return n, err
}

// IsASCIIRune reports whether the byte is an ASCII byte and not part of an UTF8 rune
// with length > 1.
func IsASCIIRune(b byte) bool { return b&0x80 == 0x00 }
