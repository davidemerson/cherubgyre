package controllers

import (
	"cherubgyre/repositories"
	"cherubgyre/services"
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

// adminToken is installed by SetAdminToken at startup. An empty value
// causes every admin request to be rejected.
var adminToken string

// SetAdminToken installs the shared-secret admin token. Called from main().
func SetAdminToken(token string) {
	adminToken = token
}

// requireAdmin returns true and writes a 401 when the request does not
// present the expected X-Admin-Token header. Rejected requests are
// recorded to the audit log so failed auth attempts show up in the
// forensic trail.
func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if adminToken == "" {
		http.Error(w, "admin disabled", http.StatusServiceUnavailable)
		repositories.WriteAudit(repositories.AuditEntry{
			Action:    "admin_auth",
			Actor:     services.ClientIP(r),
			Target:    r.URL.Path,
			RequestID: RequestIDFromContext(r.Context()),
			Result:    "failure",
			Error:     "admin disabled",
		})
		return false
	}
	got := r.Header.Get("X-Admin-Token")
	if subtle.ConstantTimeCompare([]byte(got), []byte(adminToken)) != 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		repositories.WriteAudit(repositories.AuditEntry{
			Action:    "admin_auth",
			Actor:     services.ClientIP(r),
			Target:    r.URL.Path,
			RequestID: RequestIDFromContext(r.Context()),
			Result:    "failure",
			Error:     "invalid admin token",
		})
		return false
	}
	return true
}

func AdminDeregisterUser(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]
	actor := services.ClientIP(r)
	requestID := RequestIDFromContext(r.Context())

	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	slog.Info("admin deregister request",
		slog.String("target", sanitizeForLog(username)),
		slog.String("actor_ip", actor),
		slog.String("request_id", requestID))

	err := services.DeregisterUser(username, "Admin Manual Deregistration")
	if err != nil {
		slog.Error("admin deregister failed",
			slog.String("target", sanitizeForLog(username)),
			slog.Any("err", err))
		repositories.WriteAudit(repositories.AuditEntry{
			Action:    "admin_deregister",
			Actor:     actor,
			Target:    username,
			RequestID: requestID,
			Result:    "failure",
			Error:     err.Error(),
		})
		if err.Error() == "user not found" {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	repositories.WriteAudit(repositories.AuditEntry{
		Action:    "admin_deregister",
		Actor:     actor,
		Target:    username,
		RequestID: requestID,
		Result:    "success",
	})

	w.WriteHeader(http.StatusOK)
	response := map[string]string{"message": "User successfully deregistered"}
	_ = json.NewEncoder(w).Encode(response)
}
