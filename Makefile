run:
	@./bin/kquetolk tcpServer

build:
	@go build -ldflags="-s -w" -o ./bin/kquetolk cmd/kquetolk/main.go
