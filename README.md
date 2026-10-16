# GoQueue

**A concurrent background job processing service built with Go.**

GoQueue is a lightweight background job processing system that accepts jobs over HTTP and executes them using a fixed-size worker pool. It provides a concurrency-safe in-memory job store, job lifecycle tracking, a CLI client, and a full-screen terminal UI.

GoQueue supports JSON snapshot persistence, recovery of interrupted jobs after a restart, and graceful shutdown. It can run locally or in Docker.

## Features

- **Concurrent execution** — process jobs using a fixed-size worker pool and buffered job queue.
- **Thread-safe storage** — protect shared job state using Go synchronization primitives.
- **Job lifecycle management** — track jobs through queued, running, completed, and failed states.
- **HTTP API** — submit jobs, retrieve individual jobs, and list jobs.
- **Job execution** — support `uppercase` and `reverse` operations.
- **Snapshot persistence** — save and load job state using JSON snapshots.
- **Interrupted-job recovery** — requeue jobs that were left running when the service stopped.
- **Graceful shutdown** — shut down the HTTP server, stop the worker pool, and persist job state.
- **CLI client** — submit jobs and retrieve their status from the terminal.
- **Terminal UI** — interact with the job queue through a full-screen Bubble Tea interface.
- **Docker support** — use a multi-stage build, a non-root runtime user, and a named volume for persistent data.
- **Automated checks** — run tests, race detection, static analysis, and formatting checks through development tools and GitHub Actions.

## Tech Stack

| Technology               | Purpose                             |
| ------------------------ | ----------------------------------- |
| Go                       | Core implementation and concurrency |
| `net/http`               | HTTP API                            |
| `sync`                   | Thread-safe shared state            |
| `encoding/json`          | API serialization and persistence   |
| `github.com/google/uuid` | Job identifiers                     |
| Bubble Tea               | Terminal UI framework               |
| Bubbles                  | Terminal input components           |
| Lip Gloss                | Terminal styling                    |
| Docker                   | Containerization                    |
| GitHub Actions           | Continuous integration              |

## Architecture

```text
                  ┌─────────────────────┐
                  │      Clients        │
                  │                     │
                  │  CLI  │  Terminal UI│
                  └──────────┬──────────┘
                             │ HTTP
                  ┌──────────▼──────────┐
                  │      HTTP API       │
                  │ Submit / Get / List │
                  └──────────┬──────────┘
                             │
                  ┌──────────▼──────────┐
                  │     Job Store       │
                  │  Thread-safe state  │
                  └──────────┬──────────┘
                             │
                  ┌──────────▼──────────┐
                  │   Worker Pool       │
                  │  Fixed concurrency  │
                  └──────────┬──────────┘
                             │
                  ┌──────────▼──────────┐
                  │   Job Execution     │
                  │ Uppercase / Reverse │
                  └─────────────────────┘
```

The HTTP API accepts job requests and delegates them to the worker system. Workers execute supported operations and update the job's state in the store. Clients can query the API to inspect job progress and results.

The store can save a JSON snapshot during shutdown and restore saved jobs when the server starts again.

## Project Structure

```text
goqueue/
├── .github/
│   └── workflows/
│       └── ci.yml
├── cmd/
│   ├── goqueue/
│   │   └── main.go
│   ├── goqueue-cli/
│   │   └── main.go
│   └── goqueue-tui/
│       └── main.go
├── data/
│   └── jobs.json
├── internal/
│   ├── api/
│   │   ├── handler.go
│   │   └── handler_test.go
│   ├── job/
│   │   ├── job.go
│   │   └── job_test.go
│   ├── store/
│   │   ├── store.go
│   │   └── store_test.go
│   └── worker/
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

The `data/jobs.json` file is the snapshot location used by the application. Depending on your repository setup, it may be generated at runtime rather than committed to Git.

## Getting Started

### Prerequisites

- Go version required by `go.mod`
- Git
- GNU Make, for Makefile commands
- Docker, if you want to run the containerized version

### Clone the repository

```bash
git clone https://github.com/Varun0856/goqueue.git
cd goqueue
```

### Build the applications

```bash
make build
```

This produces three binaries:

| Binary        | Purpose                        |
| ------------- | ------------------------------ |
| `goqueue`     | HTTP server                    |
| `goqueue-cli` | Command-line client            |
| `goqueue-tui` | Full-screen terminal interface |

### Start the server

```bash
make run
```

Alternatively:

```bash
go run ./cmd/goqueue
```

The server listens on `http://localhost:8080`.

Keep the server running while using the CLI or TUI.

## HTTP API

The API exposes three endpoints.

| Method | Endpoint     | Description          |
| ------ | ------------ | -------------------- |
| `POST` | `/jobs`      | Submit a new job     |
| `GET`  | `/jobs`      | List jobs            |
| `GET`  | `/jobs/{id}` | Retrieve a job by ID |

### 1. Submit a job

**Endpoint:** `POST /jobs`

```bash
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{"type":"uppercase","payload":"hello world"}'
```

Example response:

```json
{
  "ID": "d041e442-e7a5-45a8-ae1e-e3fc9811d8ea",
  "status": "queued"
}
```

The server generates a job ID and returns the initial status. A job may be picked up by a worker before the client makes its next request.

### 2. List jobs

**Endpoint:** `GET /jobs`

```bash
curl http://localhost:8080/jobs
```

This endpoint returns the collection of jobs available through the API, allowing clients to inspect the queue without knowing individual job IDs.

### 3. Retrieve a job

**Endpoint:** `GET /jobs/{id}`

```bash
curl http://localhost:8080/jobs/d041e442-e7a5-45a8-ae1e-e3fc9811d8ea
```

Example response:

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

The response reflects the job state at the time of the request. A job may still be queued or running, so poll again if you need its final result.

### Supported job types

| Type        | Behavior                   | Example input | Example output |
| ----------- | -------------------------- | ------------- | -------------- |
| `uppercase` | Converts text to uppercase | `hello world` | `HELLO WORLD`  |
| `reverse`   | Reverses the input string  | `hello world` | `dlrow olleh`  |

### Job lifecycle

```text
Queued ──► Running ──► Completed
                    └► Failed
```

Jobs enter the queue, are picked up by a worker, and eventually reach a terminal state. The store tracks their current status, payload, result, and error information.

## CLI Client

The CLI lets you submit jobs and retrieve their state without using `curl` directly.

### Submit a job

Using Make:

```bash
make run-cli ARGS='submit uppercase "hello world"'
```

Or using the compiled binary:

```bash
./goqueue-cli submit uppercase "hello world"
```

### Retrieve a job

```bash
./goqueue-cli get <job-id>
```

Or:

```bash
make run-cli ARGS='get <job-id>'
```

Replace `<job-id>` with the ID returned by the submission command.

The CLI connects to the server at `localhost:8080` in its current configuration.

## Terminal UI

GoQueue includes a full-screen terminal interface built with Bubble Tea and Lip Gloss.

### Launch the TUI

```bash
make run-tui
```

Or:

```bash
./goqueue-tui
```

Start the GoQueue server first.

### Available interactions

| Key                    | Action                          |
| ---------------------- | ------------------------------- |
| `↑` / `↓` or `j` / `k` | Navigate the job list           |
| `Enter`                | Open the selected job's details |
| `s`                    | Open the job submission form    |
| `r`                    | Refresh the current view's data |
| `Esc`                  | Return or cancel                |
| `q`                    | Quit                            |
| `Ctrl+C`               | Exit the TUI                    |

The dashboard refreshes automatically while displayed. Opening a job's details shows the most recently loaded job data; press `r` to request an updated list.

The submission form supports the `uppercase` and `reverse` operations.

## Persistence and Recovery

GoQueue uses JSON snapshots stored at `data/jobs.json`.

### Startup

1. Load the saved snapshot, if present.
2. Recover jobs that were left in the `running` state by resetting them to `queued`.
3. Start the worker pool and HTTP server.

### Shutdown

During graceful shutdown, GoQueue stops accepting new HTTP requests, stops the worker pool according to its shutdown behavior, and saves the current store as a JSON snapshot.

### Durability limitations

Snapshot persistence is **not equivalent to a transactional database or a write-ahead log**.

- Changes made after the last successful snapshot may be lost if the process crashes.
- Recovery can requeue an interrupted job, but the application cannot guarantee exactly-once execution.
- If a job's side effects occurred before a crash, re-execution may repeat those side effects.
- The JSON snapshot is intended for this project's lightweight persistence requirements, not as a substitute for a production database.

## Docker

### Build the image

```bash
make docker-build
```

Or:

```bash
docker build -t goqueue .
```

### Run with persistent storage

The Makefile's `docker-run` target mounts a named volume at `/app/data`:

```bash
make docker-run
```

Equivalent command:

```bash
docker run -d \
  --name goqueue \
  -p 8080:8080 \
  -v goqueue-data:/app/data \
  goqueue:latest
```

The named volume retains snapshot files across container replacement.

### Inspect the container

```bash
docker ps
docker logs goqueue
```

### Test the containerized API

```bash
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{"type":"reverse","payload":"hello docker"}'
```

### Recreate the container

```bash
docker stop goqueue
docker rm goqueue

docker run -d \
  --name goqueue \
  -p 8080:8080 \
  -v goqueue-data:/app/data \
  goqueue:latest
```

The volume survives container removal. However, retaining the volume does not automatically save in-memory changes: the application must successfully write its snapshot before those changes are durable.

## Development

### Make targets

| Command                   | Description                           |
| ------------------------- | ------------------------------------- |
| `make build`              | Build the server, CLI, and TUI        |
| `make test`               | Run all Go tests                      |
| `make test-race`          | Run tests with the race detector      |
| `make vet`                | Run `go vet`                          |
| `make fmt`                | List files that need formatting       |
| `make run`                | Start the server                      |
| `make run-cli ARGS='...'` | Run the CLI                           |
| `make run-tui`            | Launch the TUI                        |
| `make docker-build`       | Build the Docker image                |
| `make docker-run`         | Run the container with a named volume |
| `make clean`              | Remove generated binaries             |

### Run the test suite

```bash
make test
```

### Check for concurrency issues

```bash
make test-race
```

### Run static analysis

```bash
make vet
```

### Check formatting

```bash
make fmt
```

The `fmt` target reports files that need formatting; it does not modify them. To format Go files:

```bash
gofmt -w .
```

Before publishing changes, run:

```bash
make fmt
make test
make test-race
make vet
make build
```

## Continuous Integration

GoQueue uses GitHub Actions to automate quality checks on pushes and pull requests targeting `main`.

The workflow in `.github/workflows/ci.yml` checks compilation, static analysis, tests with the race detector, and Go formatting.

These checks help detect compilation errors, concurrency problems, and inconsistent formatting before changes are merged.

## Future Improvements

Potential extensions include:

- Configurable worker counts and queue capacity.
- Retry policies and configurable job attempts.
- Write-ahead logging or database-backed persistence.
- Delayed and scheduled jobs.
- Metrics and observability with Prometheus and Grafana.
- Docker Compose deployment.
- Distributed job processing across multiple instances.
- More job types and configurable execution handlers.

## License

GoQueue is distributed under the MIT License. See [LICENSE](LICENSE) for details.
