.PHONY: build run test race benchmark

build:
	go build -o build/engine cmd/api/main.go

run:
	go run cmd/api/main.go

test:
	go test ./internal/engine/...

race:
	go test -v -race ./internal/engine/...

benchmark:
	chmod +x benchmark.sh
	./benchmark.sh