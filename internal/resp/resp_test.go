package resp

import (
	"bytes"
	"testing"
)

func TestHandleBulkString(t *testing.T) {

	buf := []byte("$5\r\nhello\r\n")

	got, consumed, err := handleBulkString(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []byte("hello")
	if !bytes.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}

	if consumed != 11 {
		t.Fatalf("consumed %d bytes, want 11", consumed)
	}
}
