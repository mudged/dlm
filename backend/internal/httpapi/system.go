package httpapi

import (
	"encoding/json"
	"net/http"
)

// confirmPhrase is the exact string the caller must supply in the request body
// to authorise a factory reset (WI-15, REQ-017 BR-3).
const confirmPhrase = "FACTORY RESET"

// postFactoryReset wipes all application data and re-seeds default samples (REQ-017).
//
// The caller MUST send { "confirm": "FACTORY RESET" }; any other body is
// rejected with 422 (confirmation_required) before any destructive work runs.
//
// Shutdown ordering (must happen before the DB is wiped):
//  1. Routine engine — may still be mutating light state via scene patches.
//  2. Capture sweeps — driving raw LED frames on devices.
//  3. Reconstruction jobs — running Python CV children against old model IDs.
//
// After the workers are stopped the DB is reset and light state is reloaded.
// The workers are not permanently disabled; they resume accepting new work
// from the next HTTP request.
func (d *apiDeps) postFactoryReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	const maxBody = 512
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)

	var body struct {
		Confirm string `json:"confirm"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil || body.Confirm != confirmPhrase {
		writeAPIError(w, http.StatusUnprocessableEntity, "confirmation_required",
			`body must be {"confirm":"FACTORY RESET"}`)
		return
	}

	// Stop background workers before wiping the DB so they cannot write light
	// state or reference scene/model IDs that will no longer exist.
	if d.engine != nil {
		d.engine.Shutdown()
	}
	if d.capture != nil {
		d.capture.Shutdown()
	}
	if d.reconstruct != nil {
		d.reconstruct.Shutdown()
	}

	if err := d.store.FactoryReset(r.Context()); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "factory reset failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
