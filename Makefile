run:
	go run ./cmd/server

test:
	go test ./...

race:
	go test -race ./...

bench:
	go test -bench=./...