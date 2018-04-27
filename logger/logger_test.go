package logger

import (
	"io"
	"testing"
	"time"
)

var W io.Writer

func TestCreateLogger(t *testing.T) {

	l := CreateLogger("/tmp/sock")

	l.Write("test 1")
	time.Sleep(2E9)

	W = l.Writer()
}

func TestAfter(t *testing.T) {
	msg := "This is test 2 after close...test"

	_, err := W.Write([]byte(msg))
	if err != nil {
		t.Fatal("Write error:", err)

	}
}
