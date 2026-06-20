package logging

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Logger struct {
	out io.Writer
}

func New() *Logger {
	return &Logger{out: os.Stdout}
}

func NewWithWriter(w io.Writer) *Logger {
	return &Logger{out: w}
}

func (l *Logger) Printf(format string, args ...any) {
	_, _ = fmt.Fprintf(l.out, format+"\n", args...)
}

func (l *Logger) Group(name string, fn func()) {
	_, _ = fmt.Fprintf(l.out, "::group::%s\n", name)
	fn()
	_, _ = fmt.Fprintln(l.out, "::endgroup::")
}

func (l *Logger) Notice(format string, args ...any) {
	_, _ = fmt.Fprintf(l.out, "::notice::"+format+"\n", args...)
}

func (l *Logger) Warning(format string, args ...any) {
	_, _ = fmt.Fprintf(l.out, "::warning::"+format+"\n", args...)
}

func (l *Logger) Error(format string, args ...any) {
	_, _ = fmt.Fprintf(l.out, "::error::"+format+"\n", args...)
}

func Redact(s string) string {
	if s == "" {
		return s
	}
	return strings.Repeat("*", len(s))
}
