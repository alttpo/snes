package testinglogger

import (
	"strings"
	"testing"
)

type TestingLogger struct {
	testing.TB
	sb strings.Builder
}

func (l *TestingLogger) Grow(n int) {
	l.sb.Grow(n)
}

func (l *TestingLogger) Write(p []byte) (n int, err error) {
	return l.sb.Write(p)
}

func (l *TestingLogger) Commit() {
	l.TB.Log(l.sb.String())
	l.sb.Reset()
}
