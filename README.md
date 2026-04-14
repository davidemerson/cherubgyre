# cherubgyre

[![CI](https://github.com/davidemerson/cherubgyre/actions/workflows/ci.yml/badge.svg)](https://github.com/davidemerson/cherubgyre/actions/workflows/ci.yml)

**cherubgyre** is the backend for an anonymous community-defense social
network. It lets at-risk users (journalists, dissidents, organizers)
signal duress to their followers while remaining anonymous, and is
designed around a specific threat: a coercer who has physical access
to a user's phone and demands they unlock the app.

The design is documented at
[nnix.com/projects/cherubgyre](https://nnix.com/projects/cherubgyre/).
This repository is a Go HTTP service that implements the backend API
described there.

## What the backend does

- **Invite-only registration** with single-use invite codes minted by
  existing users (max 5 per user per rolling 168-hour window).
- **Two PINs per account**: a normal PIN and a **duress PIN**.
  Logging in with the duress PIN silently creates a duress signal
  visible to the user's followers, while the app UI returns
  randomized-but-plausible fake data to the coercer.
- **Bilateral follow graph** with pending/accepted states, with
  follower ban and unfollow support.
- **Duress signals** that followers can see via `/duress/following`,
  rate-limited to one user-initiated signal per hour.
- **Launch Lock**: ten wrong PIN attempts permanently delete the
  account. This is intentional — an attacker with the device but
  without the PIN cannot brute-force it. The attack surface around it
  (username enumeration, response timing) is hardened.
- **Wordlist usernames** in the spec-compliant `angel-type-city`
  format (e.g. `cherub-gyre-chicago`).
- **UUIDv7 UIDs** assigned at registration for stable identity
  across username recycling.
- **1-year inactivity deregistration** as a background sweep.

JSON persistence with atomic temp-file + rename writes and per-file
`sync.RWMutex` guards. Bcrypt (cost 12) for PIN storage. JWT (HS256)
for session tokens. No external database required.

## Requirements

- **Go 1.25+** for native development (`go` toolchain, `go mod`)
- **Docker 24+** and **Docker Compose v2** for the containerized flow
- **Python 3.11+** with `pytest` and `requests` if you want to run
  the integration test suite locally

## Quick start — Docker Compose

The easiest way to get a server running:

```bash
git clone https://github.com/davidemerson/cherubgyre.git
cd cherubgyre

cp .env.example .env
# Open .env and fill in JWT_SECRET and ADMIN_TOKEN. Generate each with:
#   openssl rand -hex 32

docker compose up --build
```

The server will be listening on `http://localhost:8080`. Verify with:

```bash
curl http://localhost:8080/health
# => The server is in good health
```

Persistence (users, followers, duress signals, used invite codes)
lives in a named Docker volume called `cherubgyre-data`, so state
survives container restarts. To wipe everything and start fresh:

```bash
docker compose down -v
```

### Building the image without Compose

```bash
docker build -t cherubgyre:local .

docker run --rm -p 8080:8080 \
  -e JWT_SECRET="$(openssl rand -hex 32)" \
  -e ADMIN_TOKEN="$(openssl rand -hex 32)" \
  -v cherubgyre-data:/data \
  --name cherubgyre \
  cherubgyre:local
```

The image runs as an unprivileged user (`cherub`), keeps persistent
state under `/data`, and exposes a healthcheck on `/health`.

## Development setup — native Go

For a fast edit/run loop without Docker:

```bash
git clone https://github.com/davidemerson/cherubgyre.git
cd cherubgyre

export JWT_SECRET="$(openssl rand -hex 32)"
export ADMIN_TOKEN="$(openssl rand -hex 32)"

go run ./
```

Or build a binary:

```bash
go build -o cherubgyre ./
./cherubgyre
```

When running natively the JSON files (`users.json`, `followers.json`,
`duress.json`, `used_invite_codes.json`) are created in the current
working directory. All four are gitignored.

### The pre-commit hook

An optional git hook is provided that runs `go vet`, `go build`, and
`go test` before each commit. Enable it once per clone:

```bash
git config core.hooksPath .githooks
```

## Environment variables

| Variable                    | Required | Default | Purpose                                                                                        |
|-----------------------------|----------|---------|------------------------------------------------------------------------------------------------|
| `JWT_SECRET`                | **yes**  | —       | HS256 signing key for session tokens. Must be **at least 32 bytes**. Server refuses to start otherwise. |
| `ADMIN_TOKEN`               | **yes**  | —       | Shared secret required as the `X-Admin-Token` header on `/admin/*`. Must be at least 16 bytes.|
| `PORT`                      | no       | `8080`  | TCP port to bind.                                                                              |
| `LOG_LEVEL`                 | no       | `info`  | `debug` / `info` / `warn` / `error`. Structured JSON logs are always enabled.                  |
| `LOGIN_RATE_LIMIT`          | no       | `10`    | Max `/login` attempts per client IP within the window.                                         |
| `LOGIN_RATE_WINDOW_SECONDS` | no       | `300`   | Sliding window length for login rate limiting, in seconds.                                     |
| `RUN_INACTIVITY_JOB`        | no       | `true`  | Set to `false` to disable the background inactivity sweep. Useful on multi-replica deployments where only one process should run it. |
| `MASTER_INVITE_CODE`        | no       | built-in | Overrides the built-in bootstrap invite code. Set to a fresh UUID to rotate the bootstrap credential without recompiling. |

## Registering your first user

The server ships with one hardcoded bootstrap invite code so you can
create the first account. After that, existing users mint invites for
new users via `GET /invite`.

```bash
# Register (use the master invite to bootstrap the first user)
curl -sX POST http://localhost:8080/register \
  -H 'Content-Type: application/json' \
  -d '{
    "normal_pin": "1234",
    "duress_pin": "9876",
    "invite_code": "4f88690e-0fbc-47b9-88e3-2d5ee2ac03d2"
  }'
# Response includes the auto-generated username like "cherub-gyre-chicago".

# Login with the normal PIN to get a session token
curl -sX POST http://localhost:8080/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"cherub-gyre-chicago","pin":"1234"}'
```

The returned token goes in the `Authorization` header (with or without
the `Bearer ` prefix) for authenticated endpoints.

## Running tests

### Unit tests (Go)

```bash
go test ./... -race
```

There are currently no `*_test.go` files in the tree; the integration
suite below is the primary verification. Adding Go-level tests is a
natural next step.

### Integration tests (pytest)

The integration suite lives at `test/test_api.py` and exercises the
live HTTP surface. Start the server in one shell, run the suite in
another:

```bash
# shell 1 — a test instance with relaxed rate limits so the rapid-
# fire suite doesn't trip the /login limiter
export JWT_SECRET="$(openssl rand -hex 32)"
export ADMIN_TOKEN="ci-test-admin-token-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export LOGIN_RATE_LIMIT=1000
export LOGIN_RATE_WINDOW_SECONDS=60
go run ./

# shell 2
pip install pytest requests
ADMIN_TOKEN=ci-test-admin-token-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
  pytest test/test_api.py -v
```

The suite covers registration, login opacity, profile duress-mode,
follow graph, duress signal posting/cancellation, duress rate
limiting, and admin authentication — around 27 cases in total.

## Continuous integration

Every push and pull request to `main` runs the full pipeline:

- `go build` and `go vet`
- `go test -race`
- `golangci-lint` (errcheck, govet, ineffassign, staticcheck, unused,
  gosec, misspell, unconvert, unparam)
- `govulncheck` against the latest patched Go stdlib
- `gosec` as a separate job
- The integration suite above, run against a freshly-built binary
  in a temp directory

GitHub's default CodeQL setup handles code scanning. Dependabot is
configured to propose weekly updates to Go modules, GitHub Actions,
and the Dockerfile base images.

## Project layout

```
cherubgyre/
├── .github/              # CI workflows + Dependabot config
├── .githooks/            # Optional pre-commit hook
├── config/               # Startup config loader (env vars)
├── controllers/          # HTTP handlers + RequireAuth middleware
├── docs/                 # Reference material, test case notes
├── dtos/                 # Wire/storage structs
├── logos/                # Brand assets
├── repositories/         # JSON persistence with atomic writes
├── services/             # Business logic, rate limiting, dummy data
│   └── wordlists/        # Embedded angel/type/city username wordlists
├── test/                 # pytest integration suite (test_api.py)
├── main.go               # HTTP entry point and route table
├── Dockerfile            # Multi-stage alpine image
├── docker-compose.yml    # Local/single-host deployment
├── .env.example          # Env var template
├── .dockerignore
├── .golangci.yml
├── go.mod / go.sum
├── CODE_OF_CONDUCT.md
├── SECURITY.md
└── LICENCE               # GNU GPL v3
```

## Security notes for operators

- `JWT_SECRET` and `ADMIN_TOKEN` must be generated per deployment and
  rotated if ever exposed. There are no defaults — the server refuses
  to start without them.
- Run behind a TLS terminator (reverse proxy or load balancer). The
  server itself speaks plain HTTP.
- The built-in `LoginLimiter` and `CheckRateLimit` are in-memory and
  per-process; for multi-replica deployments you'll want either a
  sticky-session LB or a shared rate-limit store.
- `/users/map`, `/duress/following`, and the rest of the duress
  endpoints do not leak the real follower graph when the session
  JWT was minted with a duress PIN — they return seeded, stable,
  plausible fake data instead. This is the core wrench-attack
  mitigation. If you add new endpoints that return user data, make
  sure they respect the `IsDuress` flag on the request principal.
- The one hardcoded bootstrap invite code lives in
  `repositories/user_repository.go`. Once your install base has
  its first users, consider rotating it out of the binary (future
  work: move to an env var).

See [`SECURITY.md`](SECURITY.md) for the vulnerability disclosure
policy.

## Contributing

1. Fork, branch from `main`.
2. Make your changes. The pre-commit hook will run build + vet +
   test; CI enforces the same.
3. Open a pull request with a clear description. If your change
   affects the wire format, the duress behavior, or the storage
   layer, call that out explicitly.

Please read [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md) before opening
an issue or PR.

## License

Licensed under the [GNU General Public License v3.0](LICENCE).

## Contact

Issues and contributions welcome at
[github.com/davidemerson/cherubgyre/issues](https://github.com/davidemerson/cherubgyre/issues).
