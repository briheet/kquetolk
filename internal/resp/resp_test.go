package resp

import (
	"testing"
)

func TestSimpleString(t *testing.T) {
	buf := []byte("+OK\r\n")

	v, n, err := Parse(buf)
	if err != nil {
		t.Fatal(err)
	}

	if n != len(buf) {
		t.Fatalf("expected %d bytes, got %d", len(buf), n)
	}

	if v.Type != '+' || v.Str != "OK" {
		t.Fatalf("unexpected value: %+v", v)
	}
}

func TestError(t *testing.T) {
	buf := []byte("-ERR unknown command\r\n")

	v, _, err := Parse(buf)
	if err != nil {
		t.Fatal(err)
	}

	if v.Type != '-' || v.Str != "ERR unknown command" {
		t.Fatalf("unexpected value: %+v", v)
	}
}

func TestInteger(t *testing.T) {
	buf := []byte(":1000\r\n")

	v, _, err := Parse(buf)
	if err != nil {
		t.Fatal(err)
	}

	if v.Type != ':' || v.Int != 1000 {
		t.Fatalf("unexpected value: %+v", v)
	}
}

func TestBulkString(t *testing.T) {
	buf := []byte("$5\r\nhello\r\n")

	v, n, err := Parse(buf)
	if err != nil {
		t.Fatal(err)
	}

	if n != len(buf) {
		t.Fatalf("bytes mismatch")
	}

	if v.Type != '$' || v.Str != "hello" {
		t.Fatalf("unexpected value: %+v", v)
	}
}

func TestEmptyBulkString(t *testing.T) {
	buf := []byte("$0\r\n\r\n")

	v, _, err := Parse(buf)
	if err != nil {
		t.Fatal(err)
	}

	if v.Str != "" {
		t.Fatalf("expected empty string")
	}
}

func TestNullBulkString(t *testing.T) {
	buf := []byte("$-1\r\n")

	v, _, err := Parse(buf)
	if err != nil {
		t.Fatal(err)
	}

	if !v.Nil {
		t.Fatalf("expected nil bulk string")
	}
}

func TestArrayOfBulkStrings(t *testing.T) {
	buf := []byte("*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n")

	v, _, err := Parse(buf)
	if err != nil {
		t.Fatal(err)
	}

	if v.Type != '*' || len(v.Array) != 2 {
		t.Fatalf("unexpected value: %+v", v)
	}

	if v.Array[0].Str != "hello" || v.Array[1].Str != "world" {
		t.Fatalf("unexpected array contents")
	}
}

func TestNestedArray(t *testing.T) {
	buf := []byte("*2\r\n*2\r\n:1\r\n:2\r\n+OK\r\n")

	v, _, err := Parse(buf)
	if err != nil {
		t.Fatal(err)
	}

	if len(v.Array) != 2 {
		t.Fatalf("wrong array size")
	}

	if v.Array[0].Array[0].Int != 1 {
		t.Fatalf("nested value wrong")
	}
}

func TestPartialBulkString(t *testing.T) {
	buf := []byte("$5\r\nhel")

	_, _, err := Parse(buf)
	if err != ErrIncomplete {
		t.Fatalf("expected ErrIncomplete")
	}
}

func TestPartialArray(t *testing.T) {
	buf := []byte("*2\r\n$5\r\nhello\r\n$")

	_, _, err := Parse(buf)
	if err != ErrIncomplete {
		t.Fatalf("expected ErrIncomplete")
	}
}

func TestMultipleValues(t *testing.T) {
	buf := []byte("+OK\r\n:1\r\n")

	v1, n1, err := Parse(buf)
	if err != nil {
		t.Fatal(err)
	}

	v2, _, err := Parse(buf[n1:])
	if err != nil {
		t.Fatal(err)
	}

	if v1.Str != "OK" || v2.Int != 1 {
		t.Fatalf("pipeline parse failed")
	}
}

func TestInvalidType(t *testing.T) {
	buf := []byte("?invalid\r\n")

	_, _, err := Parse(buf)
	if err != ErrProtocol {
		t.Fatalf("expected protocol error")
	}
}

func TestInvalidInteger(t *testing.T) {
	buf := []byte(":abc\r\n")

	_, _, err := Parse(buf)
	if err != ErrProtocol {
		t.Fatalf("expected protocol error")
	}
}
