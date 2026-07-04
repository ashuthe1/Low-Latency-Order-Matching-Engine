.PHONY: build run test race benchmark profile view-profile

build:
	mkdir -p build
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

profile:
	@echo "🎯 Running 20-second CPU profiling window..."
	go test -bench=. -cpuprofile=build/cpu.prof ./internal/engine/...
	@echo "✅ Profile saved to build/cpu.prof. To view, run: go tool pprof build/cpu.prof"

view-profile:
	go tool pprof build/cpu.prof