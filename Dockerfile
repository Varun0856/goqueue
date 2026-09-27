FROM golang:1.26.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./ 

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o goqueue ./cmd/goqueue/

# ------------------

FROM alpine:latest

RUN adduser -D -u 10001 appuser

WORKDIR /app

RUN mkdir -p /app/data && chown -R appuser:appuser /app/data

COPY --from=builder /app/goqueue/ ./goqueue

RUN chown appuser:appuser goqueue

USER appuser

EXPOSE 8080

CMD ["./goqueue"]
