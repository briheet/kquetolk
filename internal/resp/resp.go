package resp

import (
	"bytes"
	"errors"
)

var (
	ERRProtocol = errors.New("unsuportted protocol")
	crlf        = []byte("\r\n")
)

type Resp struct {
}

func NewResp() Resp {
	return Resp{}
}

func (r *Resp) HandleParsing(data *[]byte) ([]byte, int, error) {

	buf := *data

	switch buf[0] {
	case '$':
		return handleBulkString(buf)
	case '+':
		return handleSimpleString(buf)
	default:
		return []byte{}, 0, ERRProtocol
	}
}

func handleBulkString(buf []byte) ([]byte, int, error) {

	bytes.Contains(buf)

	return []byte{}, 0, nil
}

func handleSimpleString(buf []byte) ([]byte, int, error) {

	return []byte{}, 0, nil
}
