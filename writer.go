// Package wordwrap provide a utility to wrap text on word boundaries.
package wordwrap

import (
	"bytes"
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"
)

// String is shorthand for declaring a new default Writer instance, used to
// immediately word-wrap a string.
func String(s string, width uint) string {
	var buf bytes.Buffer
	var writer = New(&buf, width)
	writer.WriteString(s)
	return buf.String()
}

// Bytes is shorthand for declaring a new default Writer instance, used to
// immediately word-wrap a byte slice.
func Bytes(b []byte, width uint) []byte {
	var buf bytes.Buffer
	var writer = New(&buf, width)
	writer.Write(b)
	return buf.Bytes()
}

// Writer wraps UTF-8 encoded text at word boundaries when lines exceed a limit
// number of characters. Newlines are preserved, including consecutive and
// trailing newlines, though trailing whitespace is stripped from each line.
type Writer struct {
	writer      io.Writer    // default writer
	width       int          // recommended line length in runes
	tabWidh     int          // the width of tab characters
	pos         int          // curent line position
	space       bytes.Buffer // trailing word spaces
	word        bytes.Buffer // word builder
	wordLen     int          // word length in runes
	newLine     bool         // newline flag
	prefix      string       // prefix for new line
	prefixLen   int          // prefix length in runes
	breakpoints []rune       // additional word break runes
}

// New returns a new initialized wrapper over io.Writer to write lines with
// word wrap after a given position in the line.
func New(w io.Writer, width uint) *Writer {
	return &Writer{
		writer: w,
		width:  int(width),
	}
}

// SetTabWidth sets the width of tab characters.
//
// Writer attempts to handle tab characters gracefully, converting them to
// spaces aligned on the boundary. If width is 0, when used tab character as is
// by default.
func (w *Writer) SetTabWidth(width int) {
	w.tabWidh = width
}

// SetPrefix add prefix for writing on start of newline. The prefix does not
// affect the first line.
func (w *Writer) SetPrefix(s string) {
	w.prefix = s
	w.prefixLen = utf8.RuneCountInString(s)
}

// GetPrefix return the current Writer prefix.
func (w *Writer) GetPrefix() string {
	return w.prefix
}

// SetBreakpoints set additional word breakpoint runes. For exaple: "-:^".
func (w *Writer) SetBreakpoints(s string) {
	w.breakpoints = bytes.Runes([]byte(s))
}

func (w *Writer) isBreakpoint(c rune) bool {
	for _, r := range w.breakpoints {
		if r == c {
			return true
		}
	}
	return false
}

// SetPosition set current line position for correct word wrapping.
// A negative value will increase the allowable length of the first line.
func (w *Writer) SetPosition(p int) {
	w.pos = p
}

func (w *Writer) writeSpaces() error {
	w.pos += w.space.Len()
	_, err := w.space.WriteTo(w.writer)
	return err
}

func (w *Writer) writePrefix() error {
	if !w.newLine || w.prefixLen < 1 {
		return nil
	}
	w.newLine = false
	w.pos += w.prefixLen
	_, err := io.WriteString(w.writer, w.prefix)
	return err
}

func (w *Writer) writeWord() error {
	if w.wordLen < 1 {
		return nil
	}
	if err := w.writePrefix(); err != nil {
		return err
	}
	if err := w.writeSpaces(); err != nil {
		return err
	}
	_, err := w.word.WriteTo(w.writer)
	w.pos += w.wordLen
	w.wordLen = 0
	return err
}

func (w *Writer) writeNewLine() error {
	if err := w.writePrefix(); err != nil {
		return err
	}
	w.newLine = true
	w.pos = 0
	w.space.Reset()
	_, err := w.writer.Write([]byte{'\n'})
	return err
}

// Write wraps UTF-8 encoded text at word boundaries when lines exceed a limit
// number of characters. Newlines are preserved, including consecutive and
// trailing newlines, though trailing whitespace is stripped from each line.
//
// It returns the number of bytes written and any write error encountered.
func (w *Writer) Write(b []byte) (n int, err error) {
	if w.width < 1 {
		return w.writer.Write(b) // no wrap
	}
	// read all by runes
	for len(b) > 0 {
		curr, size := utf8.DecodeRune(b) // current rune
		b = b[size:]                     // skip rune from source
		n += size

		switch {
		case curr == '\n': // end of current line
			// see if we can add the content of the space buffer to the current line
			if w.wordLen == 0 {
				if w.pos+w.space.Len() > w.width {
					w.pos = 0
					w.space.Reset()
				} else {
					// preserve whitespace
					w.space.WriteTo(w.writer)
				}
			}
			w.writeWord()
			w.writeNewLine()
		case unicode.IsSpace(curr): // end of current word
			w.writeWord()
			if curr == '\t' && w.tabWidh > 0 {
				// Replace tabs with spaces while preserving alignment.
				w.space.Write(bytes.Repeat([]byte{' '}, w.tabWidh-w.pos%w.tabWidh))
			} else {
				w.space.WriteRune(curr)
			}
		case w.isBreakpoint(curr): // valid breakpoint
			w.writeSpaces()
			w.writeWord()
			// encode & write current rune
			var b = make([]byte, utf8.UTFMax)
			size := utf8.EncodeRune(b, curr)
			b = b[:size]
			w.writer.Write(b)
			w.pos++
		default: // any other character
			w.word.WriteRune(curr)
			w.wordLen++
			// add a line break if the current word would exceed the line's
			// character limit
			if w.pos+w.wordLen+w.space.Len() >= w.width &&
				w.wordLen <= w.width {
				w.writeNewLine()
			}
		}
	}
	// output last word
	w.writeWord()
	return n, err
}

// WriteString implement io.WrieString. It returns the number of bytes written
// and any write error encountered.
func (w *Writer) WriteString(str string) (n int, err error) {
	return w.Write([]byte(str))
}

// WriteByte write byte to Writer.
func (w *Writer) WriteByte(c byte) (err error) {
	_, err = w.Write([]byte{c})
	return err
}

// WriteRune write rune to Writer. It returns the number of bytes written and
// any write error encountered.
func (w *Writer) WriteRune(r rune) (n int, err error) {
	var b = make([]byte, utf8.UTFMax)
	size := utf8.EncodeRune(b, r)
	return w.Write(b[:size])
}

// Printf formats according to a format specifier and writes to Writer.
// It returns the number of bytes written and any write error encountered.
func (w *Writer) Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(w, format, a...)
}
