package controllers

import (
	"cherubgyre/repositories"
	"log/slog"
	"net/http"
)

// Health is a cheap liveness endpoint: if the process is running and
// the HTTP server is answering, it's alive. Used by orchestrators as
// a "don't kill me yet" signal. Deliberately does not touch the
// filesystem so a slow disk does not fail the liveness probe.
func Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("The server is in good health"))
}

// Ready is the readiness endpoint. Unlike Health it actually tries a
// write+read+delete probe against the data directory so that a
// container whose volume has unmounted, filled up, or gone
// read-only is removed from the load-balancer rotation promptly.
// Returns 503 on any storage failure.
func Ready(w http.ResponseWriter, r *http.Request) {
	if err := repositories.HealthCheck(); err != nil {
		slog.Error("readiness probe failed", slog.Any("err", err))
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready"))
}
