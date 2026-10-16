.PHONY: build test test-race vet fmt check run run-cli run-tui docker-build docker-run clean

build:
	go build -o goqueue ./cmd/goqueue/
	go build -o goqueue-cli ./cmd/goqueue-cli/
	go build -o goqueue-tui ./cmd/goqueue-tui/

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

run:
	go run ./cmd/goqueue/

run-cli:
	go run ./cmd/goqueue-cli/ $(ARGS)

run-tui:
	go run ./cmd/goqueue-tui/

docker-build:
	docker build -t goqueue .

docker-run:
	docker run -d --name goqueue -p 8080:8080 -v goqueue-data:/app/data goqueue:latest

clean:
	rm -f goqueue goqueue-cli goqueue-tui
