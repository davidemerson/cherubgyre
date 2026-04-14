# Cherubgyre — roadmap

This is a living document describing what has been built, what has
been intentionally deferred, and what the next reasonable pass looks
like. It exists so a new contributor (or future-you) can understand
the project's state without digging through `git log` or the design
doc at <https://nnix.com/projects/cherubgyre/>.

The source of truth for *intent* is the design doc. This file is the
source of truth for *delivery state* — what's actually in the
binary right now vs. what's still pending.

---

## Status at a glance

| Area | State |
|---|---|
| HTTP server + routing | ✅ shipping |
| Invite-only registration | ✅ shipping |
| Two-PIN auth (normal + duress) | ✅ shipping |
| Follow graph (pending/accepted) | ✅ shipping |
| Launch Lock (10 fails → deregister) | ✅ shipping, hardened |
| Silent duress on duress-PIN login | ✅ shipping |
| Randomized duress-mode dummy data | ✅ shipping |
| `POST /duress` + 1/hour rate limit | ✅ shipping |
| `/duress/cancel` requires normal PIN | ✅ shipping |
| 1-year inactivity deregistration | ✅ shipping |
| Wordlist usernames (`angel-type-city`) | ✅ shipping |
| UUIDv7 UIDs (persisted + backfilled) | ✅ shipping at storage layer only |
| Bcrypt-hashed PINs at rest | ✅ shipping |
| Admin endpoint with `X-Admin-Token` | ✅ shipping |
| Admin audit log | ✅ shipping (`audit.json`) |
| Structured JSON logging (`slog`) | ✅ shipping |
| Request ID middleware | ✅ shipping |
| Security headers middleware | ✅ shipping |
| Body-size limits + max PIN length | ✅ shipping |
| Graceful shutdown | ✅ shipping |
| Atomic repository writes + RWMutex | ✅ shipping |
| Login rate limiter (per-IP) | ✅ shipping, env-tunable |
| `/health` (liveness) + `/ready` (filesystem probe) | ✅ shipping |
| Go unit tests on critical paths | ✅ shipping (29 tests) |
| pytest integration suite | ✅ shipping (27 tests) |
| GitHub Actions CI | ✅ green on main |
| Dockerfile + docker-compose | ✅ shipping |
| Accelerometer-triggered duress | ⏳ deferred |
| Timer-based self check-in | ⏳ deferred |
| Proximity broadcast (500 m) | ⏳ deferred |
| OpenAPI 3.0.3 spec file | ⏳ deferred |
| JWT refresh tokens / revocation | ⏳ deferred |
| Pagination on follower endpoints | ⏳ deferred |
| Prometheus `/metrics` endpoint | ⏳ deferred |
| CORS | ⏳ deferred (client-dependent) |
| Full wire-DTO / storage-model split | ⏳ deferred |
| UID as primary key end-to-end | ⏳ deferred |
| Real follower banlist | ⏳ deferred |
| Migration from JSON files to a database | ⏳ deferred |

---

## What's shipping

### Security overhaul (first hardening pass)

**Commit `9e893ad`** — the big foundational security / CI pass.

- Externalized JWT signing key to `JWT_SECRET`; fail-fast on startup
  if unset or < 32 bytes. Removed the hardcoded `"your_secret_key"`.
- Migrated JWT library from the abandoned
  `github.com/dgrijalva/jwt-go` (CVE-2020-26160) to
  `github.com/golang-jwt/jwt/v5`. Signing method is explicitly
  verified as HMAC to defend against `alg=none`.
- Bcrypt-hashed PINs at rest, with an idempotent one-shot startup
  migration for legacy plaintext records.
- Admin `/admin/users/{username}` endpoint now requires the
  `X-Admin-Token` header, compared constant-time via
  `crypto/subtle.ConstantTimeCompare`.
- Removed credential-leaking log lines that were printing every
  user's invite code and username.
- `/login` hardened against username enumeration: opaque error for
  unknown-user / wrong-PIN / locked-account, plus a per-IP sliding-
  window rate limiter.
- Duress-mode dummy data replaced with seeded-random per-user output
  — stable within a session (coercer refreshing sees the same thing)
  but different across users (no install-wide pattern).
- Atomic temp-file + rename writes and a per-file `sync.RWMutex` in
  a shared `fileStore` helper, eliminating the race-on-every-write
  bug in the file-based persistence layer.
- New `RequireAuth` middleware wired via subrouter; every authed
  handler reads identity from request context via `controllers.Identity(r)`.
- Username format aligned to the spec's three-word slug
  (`angel-type-city`, e.g. `cherub-gyre-chicago`) via embedded
  wordlists.
- UUIDv7 UIDs assigned at registration with an idempotent backfill
  for legacy users.
- `/duress/cancel` requires the normal PIN in the body (per spec).
- Duress signals rate-limited to 1/hour per user (per spec).
- Full GitHub Actions CI pipeline: build+vet, `go test -race`,
  `golangci-lint`, `govulncheck`, `gosec`, integration suite against
  a booted binary. Dependabot + CodeQL configured. `.golangci.yml`
  with 10 linters enabled.
- Committed binaries (`main`, `cherubgyre.exe`) untracked; `.gitignore`
  expanded; pre-commit hook rewritten to not write artifacts.

### Docker + README rewrite (commit `64aadeb`)

- Multi-stage `Dockerfile`: `golang:1.26-alpine` builder → `alpine:3.22`
  runtime with `ca-certificates`, `tini` as PID 1, `wget` for the
  healthcheck. Non-root `cherub` user with `/data` as WORKDIR and
  VOLUME.
- `docker-compose.yml` for local dev: named `cherubgyre-data` volume,
  `env_file: .env`, healthcheck wired to `/health` (later switched
  to `/ready`).
- `.env.example` template with inline `openssl rand -hex 32`
  instructions.
- Expanded `.dockerignore`.
- Full README rewrite: removed the external Docker Hub image link
  and the external test-repo link (tests live here now), added
  quick-start + native-dev flows, env var table, first-user curl
  walkthrough, project layout, security notes for operators, CI
  summary.
- Untracked the committed `users.json` test data with plaintext
  PINs; deleted the Heroku `Procfile`.

### Follow-up audit pass (second hardening pass)

**Commit `a73d4f5` — Phase 1 P0 bugs.**

- **Launch Lock persistence.** Previously the failed-attempt
  counter was never persisted on the 10th attempt, and a failed
  `DeregisterUser` write was silently discarded. A disk-full
  condition defeated Launch Lock entirely. The counter is now
  persisted on every failed attempt *before* any lock decision,
  and deregister failures return a server error instead of the
  opaque credentials error.
- **Orphan follower cleanup.** `DeleteUser` now invokes a new
  `repositories.DeleteUserRelations(username)` that strips every
  row in `followers.json` where the user appears on either side.
  Without this, username recycling would let a new user inherit
  the old user's follower graph.
- **Graceful shutdown.** `main.go` now uses `signal.NotifyContext`
  and `srv.Shutdown(ctx)` with a 30-second deadline; background
  goroutines honor the cancelable context.
- **Body-size limits.** New `MaxBodyBytes` middleware caps every
  request at 64 KiB globally; `/register`, `/validate-invite`, and
  `/login` get a tighter 8 KiB cap. Blocks the 100 MB-pin DoS.
- **Max PIN length.** `ValidatePin` now enforces a 64-byte upper
  bound.
- **Security headers middleware.** Sets HSTS, nosniff, frame-DENY,
  referrer-no-referrer, and cache-no-store on every response.

**Commit `5c72641` — Phase 2 quality cleanup.**

- **Typed sentinel errors.** Exported error values
  (`ErrIncorrectCurrentPin`, `ErrPinsMustDiffer`,
  `ErrDuressRateLimited`, etc.) and migrated controllers from
  `switch err.Error()` to `errors.Is`, so renaming an error message
  no longer silently changes an HTTP status code.
- **Dead migration removed.** `services.MigratePinHashes` and its
  startup call are gone; the plaintext fallback in
  `ValidateUserCredentials` is gone. The bcrypt path is now the only
  path.
- **`interface{}` → `any`** across 12 files.
- **Lock ordering doc** — package-level comment in
  `repositories/user_repository.go` spelling out that `userStore`
  must be acquired before `usedInviteStore`.

**Commit `0ccede8` — Phase 2 observability.**

- **`log/slog` structured logging.** 70+ stdlib `log.Printf`
  call sites across 16 files rewritten as structured
  `slog.Info` / `slog.Warn` / `slog.Error` / `slog.Debug` with
  typed fields (`slog.String`, `slog.Int`, `slog.Any("err", err)`).
  A JSON handler is installed as the process-wide default in
  `main.go`; `LOG_LEVEL` env var controls verbosity.
- **Login-attempt leak closed.** The noisy per-request "Login
  attempt for user:" line now emits only at `slog.Debug`, so
  production logs no longer differentiate unknown-user from
  wrong-PIN — aligning the log stream with the response-side
  enumeration hardening.
- **Request ID middleware.** New `controllers.RequestID` trusts
  an incoming `X-Request-ID` header (sanitized, length-capped)
  or mints a fresh UUIDv4; echoes it on the response header;
  exposes it via `controllers.RequestIDFromContext(ctx)` for any
  handler or service that wants to include it in logs.
- **Admin audit log.** New `repositories/audit_repository.go`
  with a `fileStore`-backed `audit.json`. Admin auth failures
  and admin deregister success/failure rows are persisted with
  timestamp, action, actor IP, target, request ID, result, and
  any error message.

**Commit `8031eaf` — Phase 3 ops and Go unit tests.**

- **Split `/health` from `/ready`.** `/health` stays passive
  (liveness: the process answers HTTP). `/ready` calls a new
  `repositories.HealthCheck()` that writes + syncs + reads +
  deletes a probe file in the data directory, so a container
  with a full or unmounted `/data` is removed from the
  load-balancer rotation promptly. Dockerfile `HEALTHCHECK` and
  the docker-compose healthcheck both now use `/ready`.
- **29 Go unit tests** across 5 new files:
  - `services/rate_limit_test.go` — IPLimiter allow/block,
    window expiry, IP isolation, `ClientIP` fallback chain
    including comma-split and whitespace-stripping.
  - `services/wordlists_test.go` — non-empty, clean entries,
    no duplicates, `UsernameCombinations >= 1M`.
  - `services/dummydata_test.go` — stability per seed,
    distinctness across seeds, shape, count bounds,
    `GetDummyInviteCode` is a valid UUID and changes per call.
  - `services/auth_service_test.go` — `ValidatePin` bounds,
    HS256 accept, expired rejection, wrong-signature rejection,
    **`alg=none` rejection** (explicit regression test against
    the CVE-2020-26160 class of attacks), malformed token
    rejection, `GetUsernameFromToken`, `IsDuressToken`,
    `Bearer `/`bearer ` prefix stripping.
  - `repositories/storage_test.go` — load of missing/empty
    file, round-trip, no temp-file leaks after 5 saves,
    **30-goroutine × 5-save concurrent stress test** (150
    total appends, must decode to exactly 150 entries under
    `-race`), invalid-JSON rejection, nested-struct round-
    trip, `HealthCheck` cleanup.

**Commit `9cf3c59` — Phase 4 cleanup.**

- **`MASTER_INVITE_CODE` env var.** The bootstrap invite code is
  now an overridable package variable, not a `const`. Operators
  can rotate it without recompiling; setting it to the empty
  string disables the master-code path entirely.

**Commit `03d36d8` — CI fix.** `#nosec G304` annotation on the
HealthCheck probe read (path comes from `os.CreateTemp`, not
attacker input).

---

## What's explicitly deferred

Each of these is known work. None are blocking a real user today.
Each deserves its own focused plan when it becomes the priority.

### Spec features not yet implemented

- **Accelerometer-triggered duress.** Client-side shake detection
  posting to `/duress` with `duress_type: "accelerometer"`, plus
  a backend 1-minute confirmation window and auto-signal on no
  confirm. Backend currently accepts the `duress_type` string
  but does not branch on it.
- **Timer-based self check-in.** User arms a countdown ("alert
  my followers if I don't check in within 5 minutes"). Requires
  new endpoint group `/checkin/{arm,confirm,dismiss}`, a new
  `check_ins.json` store, and a per-minute background job. None
  of this exists yet.
- **Proximity broadcast (500 m radius).** Requires adding
  location fields to the user record and a geospatial query
  layer. Strongly favors migrating to SQLite + R*Tree or
  PostgreSQL + PostGIS before implementing.
- **Map visualization.** The `/users/map` route exists but
  returns only the caller's own duress signal — misnamed. A
  real "map of nearby duress" view depends on proximity
  broadcast above.

### Ops and observability

- **Pagination on `/followers/{u}`, `/following`, `/follow/requests`.**
  Currently returns an unbounded JSON array. Not fixing an active
  problem under the current threat model, but a ceiling exists.
  Shape change would be `{"items": [...], "total": N, "offset": M,
  "limit": L}`. Wire-contract break — bundle with other wire changes.
- **Prometheus `/metrics` endpoint.** Gated behind a
  `METRICS_ENABLED` env var. Useful once there are multi-replica
  deployments or a real ops team monitoring production. Adds
  `github.com/prometheus/client_golang` as a dependency.
- **Full `context.Context` plumbing.** Currently only plumbed at
  the background-job boundary where cancellation actually matters.
  Deeper plumbing through every service → repository call is
  cosmetic until a real per-request deadline use case appears.
- **CORS.** Blocked on deciding whether a browser client is ever
  in scope. Mobile clients do not need it.

### Auth / session model

- **Refresh tokens + revocation list.** JWTs are currently 24h
  fixed expiry with no logout. Acceptable for the threat model;
  wrong for UX. A leaked token is valid until expiry.
- **`/logout` endpoint.** Requires the revocation list above.

### Refactors

- **Full wire-DTO / storage-model split.** `dtos.RegisterDTO`
  doubles as the persisted record. Low pain with the current
  convention ("plaintext PIN fields are incoming-only, never
  persisted"), but a clean split would remove the convention's
  dependence on discipline.
- **UID as primary key end-to-end.** UUIDv7 UIDs are persisted
  and backfilled, but controllers and services still key
  everything by username. Flipping the primary key is a large
  ripple that deserves its own pass.
- **Real follower banlist.** `BanFollower` currently aliases
  `RemoveFollower`; a banned follower can immediately re-request.
  Small feature, not urgent.
- **Rename `GetDuressMap` → `GetMyDuress`, `/users/map` →
  `/duress/me`.** Service exists (`GetMyDuress`) but the legacy
  route still works. A rename is a wire-contract break; bundle
  with pagination.

### Documentation

- **OpenAPI 3.0.3 spec file.** The design doc says the API is
  "OpenAPI 3.0.3/Swagger documented" but no `openapi.yaml` exists.
  Unblocks client SDK generation.

### Architecture

- **Migration from file-JSON to a real database.** Single biggest
  quality-of-life improvement available. Would eliminate most of
  `repositories/storage.go`, unlock pagination without full
  re-reads, enable full-text search and proximity queries, and
  remove the single-instance-only constraint on the inactivity
  sweep and in-memory rate limiter. SQLite is enough for a
  single-instance deployment; PostgreSQL if multi-replica is ever
  in scope.

---

## Suggested next pass

If I were picking the next focused plan, I'd prioritize in this
order:

1. **Real follower banlist** and **wire-contract cleanup pass**
   (pagination + `/users/map` rename) as a single commit that
   bumps the API version. Cheap, breaks nothing already
   production-critical, gets the wire shape to a more stable
   place.
2. **Timer-based self check-in.** The only greenfield spec feature
   that doesn't depend on a DB migration. Large but self-contained.
3. **Database migration to SQLite.** Unlocks everything else in
   the deferred list: proximity broadcast, pagination-by-query,
   multi-replica rate limiting, background-job leader election,
   audit log retention, and so on. Biggest single leverage point.
4. **OpenAPI spec.** Pure documentation, unblocks external SDK
   generation.
5. **Accelerometer / proximity.** These only become cheap after
   the DB migration.

---

## How to contribute to this roadmap

- When you ship something from the "deferred" list, move its row
  in the "Status at a glance" table to ✅ shipping and add a
  short entry under "What's shipping" with the commit SHA.
- When you discover a new deferred item, add it under "What's
  explicitly deferred" with a one-paragraph rationale — why it
  isn't blocking, what it would take to ship it, what it depends
  on.
- When priorities change, update "Suggested next pass" and say
  so in the commit.

The goal is that any single commit keeps this file coherent with
the state of `main`.
