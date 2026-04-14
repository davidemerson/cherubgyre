# syntax=docker/dockerfile:1

# Build stage — use a patched Go toolchain so govulncheck stays clean.
FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Cross-compile a static linux/amd64 binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main main.go

# Runtime stage — minimal Ubuntu with just TLS roots for outbound HTTPS.
FROM ubuntu:24.04

WORKDIR /root/

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/main .

EXPOSE 8080

ENV PORT=8080

CMD ["./main"]
