.PHONY: build test test-race vet fmt run run-cli docker-build docker-run clean

build:
	go build -o goqueue ./cmd/goqueue/
	go build -o goqueue-cli ./cmd/goqueue-cli/

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -l .

run:
	go run ./cmd/goqueue/

run-cli:
	go run ./cmd/goqueue-cli/ $(ARGS)

docker-build:
	docker build -t goqueue .

docker-run:
	docker run -d --name goqueue -p 8080:8080 -v goqueue-data:/app/data goqueue:latest

clean:
	rm -f goqueue goqueue-cli
