# syntax=docker/dockerfile:1.7
#
# Multi-stage build for cherubgyre. Produces a small alpine-based image
# running the server as an unprivileged user, with /data as a writable
# volume for the JSON persistence files.
#
# Build locally:
#   docker build -t cherubgyre:local .
#
# Or via docker compose (recommended for development):
#   cp .env.example .env   # then fill in JWT_SECRET and ADMIN_TOKEN
#   docker compose up --build

# ---------- builder ----------
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src

# Pre-download modules so repeated source edits hit the build cache.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Fully static linux binary (CGO off) so the runtime image can be
# any minimal distro without libc concerns. -trimpath and -s/-w
# strip absolute paths and debug symbols to keep the binary small
# and reproducible.
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -trimpath -o /out/cherubgyre ./

# ---------- runtime ----------
FROM alpine:3.24

# ca-certificates so outbound HTTPS (e.g. to the Dicebear avatar
# service) works. tini is used as PID 1 so SIGTERM from `docker stop`
# is forwarded cleanly to the Go server. wget is used by the
# HEALTHCHECK below.
RUN apk add --no-cache ca-certificates tini wget \
    && addgroup -S cherub \
    && adduser -S -G cherub -h /data cherub \
    && mkdir -p /data \
    && chown cherub:cherub /data

COPY --from=builder /out/cherubgyre /usr/local/bin/cherubgyre

USER cherub:cherub
WORKDIR /data
VOLUME ["/data"]

EXPOSE 8080
ENV PORT=8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/ready >/dev/null 2>&1 || exit 1

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/cherubgyre"]
