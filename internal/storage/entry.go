package storage

import "time"

type Entry struct {
	Value     []byte
	ExpiresAt int64 // Unix nano, 0 = no expiration
}

func (e *Entry) IsExpired() bool {
	return e.ExpiresAt != 0 && time.Now().UnixNano() > e.ExpiresAt
}
