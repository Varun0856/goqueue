# GoQueue

**A concurrent background job processing service written in Go.**

GoQueue is a lightweight background job processing system that accepts jobs through an HTTP API, executes them using a fixed-size worker pool, and stores job state in a concurrency-safe in-memory store with JSON snapshot persistence.

It supports job submission, status tracking, graceful shutdown, and recovery of interrupted jobs. It can run locally or inside a Docker container.

---

## Features

- **Concurrent job processing** using a fixed-size worker pool.
- **Thread-safe in-memory storage** protected by mutexes.
- **Job lifecycle management** with validated state transitions.
- **HTTP API** for submitting jobs and retrieving their status.
- **Job execution** supporting `uppercase` and `reverse` operations.
- **JSON snapshot persistence** for saving and loading job data.
- **Recovery** that requeues jobs left in the `running` state after a restart.
- **Graceful shutdown** that stops the HTTP server, drains the worker pool, and saves job state.
- **CLI client** for submitting jobs and retrieving their status.
- **Docker support** using a multi-stage build and a non-root runtime user.
- **Makefile** for common development, testing, and Docker commands.
- **GitHub Actions CI** for automated build, testing, static analysis, and formatting checks.

---

## Tech Stack

- Go
- Standard library (`net/http`, `sync`, `encoding/json`, `os`, etc.)
- UUID generation with `github.com/google/uuid`
- Docker
- GitHub Actions

---

## Project Structure

```text
.
├── .github
│   └── workflows
│       └── ci.yml
├── cmd
│   ├── goqueue
│   │   └── main.go
│   └── goqueue-cli
│       └── main.go
├── data
│   └── jobs.json
├── internal
│   ├── api
│   │   ├── handler.go
│   │   └── handler_test.go
│   ├── job
│   │   ├── job.go
│   │   └── job_test.go
│   ├── store
│   │   ├── store.go
│   │   └── store_test.go
│   └── worker
│       ├── execute.go
│       ├── execute_test.go
│       ├── pool.go
│       └── pool_test.go
├── .dockerignore
├── .gitignore
├── Dockerfile
├── go.mod
├── go.sum
├── LICENSE
├── Makefile
└── README.md
```

---

## Getting Started

### Prerequisites

- Go 1.26.1 or later
- Git
- Make (for Makefile commands)

### Clone the repository

```bash
git clone https://github.com/Varun0856/goqueue.git
cd goqueue
```

### Build the server and CLI

```bash
make build
```

This builds two binaries:

- `goqueue` — the HTTP server.
- `goqueue-cli` — the command-line client.

### Run the server

```bash
make run
```

Alternatively, run the server directly:

```bash
go run ./cmd/goqueue
```

The server listens on:

```text
http://localhost:8080
```

---

## API Reference

### 1. Submit a Job

**Endpoint:** `POST /jobs`

Submit a job by providing its type and payload.

**Request:**

```bash
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{"type":"uppercase","payload":"hello world"}'
```

**Example response:**

```json
{
  "ID": "d041e442-e7a5-45a8-ae1e-e3fc9811d8ea",
  "status": "queued"
}
```

The response contains the generated job ID and its initial status.

### 2. Get Job Status

**Endpoint:** `GET /jobs/{id}`

Retrieve the current state and result of a job.

```bash
curl http://localhost:8080/jobs/d041e442-e7a5-45a8-ae1e-e3fc9811d8ea
```

**Example response:**

```json
{
  "ID": "d041e442-e7a5-45a8-ae1e-e3fc9811d8ea",
  "Type": "uppercase",
  "Payload": "hello world",
  "Status": "completed",
  "Result": "HELLO WORLD",
  "Error": ""
}
```

The job's status may initially be `queued` or `running` before reaching a terminal state.

### Job Types

| Type        | Description                       | Example result |
| ----------- | --------------------------------- | -------------- |
| `uppercase` | Converts the payload to uppercase | `HELLO WORLD`  |
| `reverse`   | Reverses the payload              | `dlrow olleh`  |

### Job Lifecycle

```text
Queued → Running → Completed
                  ↘ Failed
```

Jobs begin in the `queued` state, transition to `running` when a worker picks them up, and finish as either `completed` or `failed`.

---

## CLI Client

GoQueue includes a command-line client for interacting with the HTTP API.

The CLI connects to the GoQueue server running at `localhost:8080`.

### Build the CLI

```bash
make build
```

Alternatively:

```bash
go build -o goqueue-cli ./cmd/goqueue-cli
```

### Submit a Job

Using the Makefile:

```bash
make run-cli ARGS='submit uppercase "hello world"'
```

Or run the binary directly:

```bash
./goqueue-cli submit uppercase "hello world"
```

### Retrieve a Job

Using the Makefile:

```bash
make run-cli ARGS='get <job-id>'
```

Or run the binary directly:

```bash
./goqueue-cli get <job-id>
```

Replace `<job-id>` with the ID returned when the job was submitted.

---

## Persistence and Recovery

GoQueue uses JSON snapshots to preserve job state across graceful restarts.

### How it works

1. The server loads previously saved jobs from `data/jobs.json` during startup.
2. Jobs left in the `running` state are reset to `queued` for recovery.
3. Workers process queued jobs.
4. During graceful shutdown, the worker pool drains and the store saves a JSON snapshot.

### Persistence limitation

Snapshots are written during graceful shutdown. A sudden crash or forced termination may cause changes made since the last snapshot to be lost.

This is snapshot-based persistence, not a transactional database or a write-ahead log.

---

## Docker

GoQueue can be built and run as a Docker container.

### Build the image

```bash
make docker-build
```

Alternatively:

```bash
docker build -t goqueue .
```

### Run the container

```bash
make docker-run
```

Alternatively:

```bash
docker run -d \
  --name goqueue \
  -p 8080:8080 \
  goqueue:latest
```

### Check container status

```bash
docker ps
```

### View logs

```bash
docker logs goqueue
```

### Test the API

```bash
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{"type":"reverse","payload":"hello docker"}'
```

### Persistent storage with a named volume

Create a Docker volume:

```bash
docker volume create goqueue-data
```

Run the container with the volume mounted:

```bash
docker run -d \
  --name goqueue \
  -p 8080:8080 \
  -v goqueue-data:/app/data \
  goqueue:latest
```

The volume is mounted at `/app/data`, where GoQueue stores its JSON snapshot.

To recreate the container while retaining the volume:

```bash
docker stop goqueue
docker rm goqueue

docker run -d \
  --name goqueue \
  -p 8080:8080 \
  -v goqueue-data:/app/data \
  goqueue:latest
```

The named volume retains the snapshot across container replacement. The application must save a snapshot before the container is removed for the latest changes to be preserved.

---

## Development

GoQueue includes a Makefile for common development tasks.

| Command                   | Description                        |
| ------------------------- | ---------------------------------- |
| `make build`              | Build the server and CLI binaries  |
| `make test`               | Run all tests                      |
| `make test-race`          | Run tests with the race detector   |
| `make vet`                | Run `go vet`                       |
| `make fmt`                | List Go files that need formatting |
| `make run`                | Run the server                     |
| `make run-cli ARGS='...'` | Run the CLI with arguments         |
| `make docker-build`       | Build the Docker image             |
| `make docker-run`         | Run the Docker container           |
| `make clean`              | Remove generated binaries          |

### Testing

Run the full test suite:

```bash
make test
```

Run the race detector:

```bash
make test-race
```

Run static analysis:

```bash
make vet
```

Check formatting:

```bash
make fmt
```

The formatting command lists files that need formatting. To format Go files, run:

```bash
gofmt -w .
```

---

## Continuous Integration

GoQueue uses **GitHub Actions** to automate checks on pushes and pull requests targeting the `main` branch.

The workflow is defined in `.github/workflows/ci.yml` and performs the following checks:

- **Build:** Verifies that the Go packages compile.
- **Vet:** Runs `go vet` for static analysis.
- **Tests:** Runs the test suite with the race detector enabled.
- **Formatting:** Checks that Go files are formatted with `gofmt`.

These checks help catch build errors, potential concurrency issues, and formatting problems before changes are merged.

---

## Future Improvements

- Retry policies and configurable job attempts.
- Persistent queue storage with a write-ahead log.
- Configurable worker pool size.
- Job scheduling and delayed execution.
- Metrics and observability with Prometheus and Grafana.
- Docker Compose deployment.
- Distributed job processing across multiple instances.

---

## License

See the [LICENSE](LICENSE) file for license details.
