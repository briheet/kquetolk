package client

import "testing"

func BenchmarkConn(b *testing.B) {

	b.ReportAllocs()

	for b.Loop() {
		var items = make([]Conn, 10_000_000)

		var n int
		for i := 0; i < len(items); i++ {
			n += len(items[i].readBuf)
			n += len(items[i].writeBuf)
		}

		// Prevent optimization
		if n == 42 {
			b.Fatal("impossible")
		}
	}

}
