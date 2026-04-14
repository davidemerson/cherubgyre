package repositories

import (
	"log/slog"
	"time"
)

// AuditEntry is one row in the admin audit log. New entries are only
// ever appended to the end of audit.json; the file is rewritten atomically
// via fileStore.saveLocked, so readers always see either the old or new
// state, never a torn write.
type AuditEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Actor     string    `json:"actor"`            // who performed the action (IP, or "system")
	Target    string    `json:"target"`           // what it was performed on (usually a username)
	RequestID string    `json:"request_id,omitempty"`
	Result    string    `json:"result"`           // "success" or "failure"
	Error     string    `json:"error,omitempty"`  // populated when Result == "failure"
}

var auditStore = newFileStore("audit.json")

// WriteAudit appends a single audit entry. Errors are logged but not
// propagated to the caller — an audit-log write failure should not
// block an otherwise-successful admin action, and the slog.Error
// call makes any failure visible to operators.
func WriteAudit(entry AuditEntry) {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	auditStore.mu.Lock()
	defer auditStore.mu.Unlock()

	var entries []AuditEntry
	if err := auditStore.loadLocked(&entries); err != nil {
		slog.Error("audit log load failed",
			slog.String("action", entry.Action), slog.Any("err", err))
		return
	}
	entries = append(entries, entry)
	if err := auditStore.saveLocked(entries); err != nil {
		slog.Error("audit log save failed",
			slog.String("action", entry.Action), slog.Any("err", err))
	}
}
