package resp

import (
	"bytes"
	"errors"
	"strconv"
)

var (
	ErrProtocol   = errors.New("resp: protocol error")
	ErrIncomplete = errors.New("resp: incomplete data")
	crlf          = []byte("\r\n")
)

type Value struct {
	Type  byte
	Str   string
	Int   int64
	Array []Value
	Nil   bool
}

func Parse(buf []byte) (Value, int, error) {
	if len(buf) == 0 {
		return Value{}, 0, ErrIncomplete
	}

	switch buf[0] {
	case '+':
		return parseSimpleString(buf)
	case '-':
		return parseError(buf)
	case ':':
		return parseInteger(buf)
	case '$':
		return parseBulkString(buf)
	case '*':
		return parseArray(buf)
	default:
		return Value{}, 0, ErrProtocol
	}
}

func readLine(buf []byte) ([]byte, int, error) {
	i := bytes.Index(buf, crlf)
	if i == -1 {
		return nil, 0, ErrIncomplete
	}
	return buf[:i], i + 2, nil
}

func parseSimpleString(buf []byte) (Value, int, error) {
	line, n, err := readLine(buf)
	if err != nil {
		return Value{}, 0, err
	}

	return Value{
		Type: '+',
		Str:  string(line[1:]),
	}, n, nil
}

func parseError(buf []byte) (Value, int, error) {
	line, n, err := readLine(buf)
	if err != nil {
		return Value{}, 0, err
	}

	return Value{
		Type: '-',
		Str:  string(line[1:]),
	}, n, nil
}

func parseInteger(buf []byte) (Value, int, error) {
	line, n, err := readLine(buf)
	if err != nil {
		return Value{}, 0, err
	}

	v, err := strconv.ParseInt(string(line[1:]), 10, 64)
	if err != nil {
		return Value{}, 0, ErrProtocol
	}

	return Value{
		Type: ':',
		Int:  v,
	}, n, nil
}

func parseBulkString(buf []byte) (Value, int, error) {
	line, n, err := readLine(buf)
	if err != nil {
		return Value{}, 0, err
	}

	length, err := strconv.Atoi(string(line[1:]))
	if err != nil {
		return Value{}, 0, ErrProtocol
	}

	if length == -1 {
		return Value{Type: '$', Nil: true}, n, nil
	}

	total := n + length + 2
	if len(buf) < total {
		return Value{}, 0, ErrIncomplete
	}

	data := buf[n : n+length]

	return Value{
		Type: '$',
		Str:  string(data),
	}, total, nil
}

func parseArray(buf []byte) (Value, int, error) {
	line, n, err := readLine(buf)
	if err != nil {
		return Value{}, 0, err
	}

	count, err := strconv.Atoi(string(line[1:]))
	if err != nil {
		return Value{}, 0, ErrProtocol
	}

	if count == -1 {
		return Value{Type: '*', Nil: true}, n, nil
	}

	values := make([]Value, 0, count)
	consumed := n

	for i := 0; i < count; i++ {
		v, c, err := Parse(buf[consumed:])
		if err != nil {
			return Value{}, 0, err
		}
		values = append(values, v)
		consumed += c
	}

	return Value{
		Type:  '*',
		Array: values,
	}, consumed, nil
}
