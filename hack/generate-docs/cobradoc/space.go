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
	start := 0
	if w.spaces > 0 {
	loop:
		for i := 0; i < len(p); i++ {
			switch p[i] {
			case ' ':
				w.spaces++
				start++
			case '\n':
				// discard spaces
				w.spaces = 0
				break loop
			default:
				_, err := w.Writer.Write([]byte(fmt.Sprintf(fmt.Sprintf("%%%ds", w.spaces), " ")))
				if err != nil {
					return 0, err
				}
				w.spaces = 0
				break loop
			}
		}
	}

	for i := start; i < len(p); i++ {
		b := p[i]
		if IsASCIIRune(b) {
			switch b {
			case ' ':
				w.spaces++
			case '\n':
				if w.spaces > 0 {
					// discard spaces
					written, err := w.Writer.Write(p[start : i-w.spaces])
					w.spaces = 0
					if err != nil {
						return start + written, err
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
	if start+w.spaces < len(p) {
		written, err := w.Writer.Write(p[start : len(p)-w.spaces])
		return start + written + w.spaces, err
	}
	return len(p), nil
}

// IsASCIIRune reports whether the byte is an ASCII byte and not part of an UTF8 rune
// with length > 1.
func IsASCIIRune(b byte) bool { return b&0x80 == 0x00 }
