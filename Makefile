run:
	go run cmd/api/main.go

test:
	go test ./...

race:
	go test -race ./...

bench:
	go test -bench=./...