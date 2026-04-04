package httpapi

import (
	"io"
	"net/http"
)

// postFactoryReset wipes all application data and re-seeds default samples (REQ-017).
func (d *apiDeps) postFactoryReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	const maxBody = 4096
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	_, _ = io.Copy(io.Discard, r.Body)
	_ = r.Body.Close()

	if err := d.store.FactoryReset(r.Context()); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "factory reset failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
