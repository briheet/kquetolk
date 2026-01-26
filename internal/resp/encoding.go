package resp

import "strconv"

func EncodeSimpleString(s string) []byte {
	return []byte("+" + s + "\r\n")
}

func EncodeError(s string) []byte {
	return []byte("-" + s + "\r\n")
}

func EncodeBulkString(s string) []byte {
	return []byte("$" + strconv.Itoa(len(s)) + "\r\n" + s + "\r\n")
}

func EncodeNullBulkString() []byte {
	return []byte("$-1\r\n")
}

func EncodeInteger(n int64) []byte {
	return []byte(":" + strconv.FormatInt(n, 10) + "\r\n")
}

func EncodeArray(items []string) []byte {
	buf := []byte("*" + strconv.Itoa(len(items)) + "\r\n")
	for _, item := range items {
		buf = append(buf, EncodeBulkString(item)...)
	}
	return buf
}
