package testinglogger

import (
	"testing"
	"unsafe"
)

type TestingLogger struct {
	testing.TB
	buf []byte
}

func (l *TestingLogger) Reserve(n int) {
	if cap(l.buf) >= n {
		return
	}

	newbuf := make([]byte, len(l.buf), n)
	copy(newbuf, l.buf)
	l.buf = newbuf
}

func (l *TestingLogger) Write(p []byte) (n int, err error) {
	l.buf = append(l.buf, p...)
	return len(p), nil
}

func (l *TestingLogger) Commit() {
	line := *(*string)(unsafe.Pointer(&l.buf))
	l.TB.Log(line)
	l.buf = l.buf[:0]
}
