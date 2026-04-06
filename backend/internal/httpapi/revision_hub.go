package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"example.com/dlm/backend/internal/store"
)

// RevisionHub broadcasts monotonic revision numbers for model and scene light-state
// topics (REQ-029 / architecture §3.18). Safe for concurrent use.
type RevisionHub struct {
	mu   sync.Mutex
	seq  map[string]uint64
	subs map[string][]chan uint64
}

// NewRevisionHub constructs an empty hub.
func NewRevisionHub() *RevisionHub {
	return &RevisionHub{
		seq:  make(map[string]uint64),
		subs: make(map[string][]chan uint64),
	}
}

func modelTopic(modelID string) string { return "m:" + modelID }
func sceneTopic(sceneID string) string { return "s:" + sceneID }

// NotifyModelLightsChanged bumps the model topic and every scene that contains this model.
func (h *RevisionHub) NotifyModelLightsChanged(ctx context.Context, st *store.Store, modelID string) {
	if h == nil || st == nil || modelID == "" {
		return
	}
	h.bump(modelTopic(modelID))
	scenes, err := st.ListSceneIDsForModel(ctx, modelID)
	if err != nil {
		return
	}
	for _, sid := range scenes {
		h.bump(sceneTopic(sid))
	}
}

// NotifyAfterSceneLightPatch notifies subscribers for each distinct model touched by a scene bulk patch.
func (h *RevisionHub) NotifyAfterSceneLightPatch(ctx context.Context, st *store.Store, states []store.ScenePatchedState) {
	if h == nil || len(states) == 0 {
		return
	}
	seen := make(map[string]bool)
	for _, ps := range states {
		mid := ps.ModelID
		if mid == "" || seen[mid] {
			continue
		}
		seen[mid] = true
		h.NotifyModelLightsChanged(ctx, st, mid)
	}
}

func (h *RevisionHub) bump(key string) uint64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.seq[key]++
	v := h.seq[key]
	for _, ch := range h.subs[key] {
		select {
		case ch <- v:
		default:
		}
	}
	return v
}

func (h *RevisionHub) subscribe(key string) (last uint64, ch <-chan uint64, unsub func()) {
	c := make(chan uint64, 32)
	h.mu.Lock()
	last = h.seq[key]
	h.subs[key] = append(h.subs[key], c)
	h.mu.Unlock()
	return last, c, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		lst := h.subs[key]
		out := lst[:0]
		for _, x := range lst {
			if x != c {
				out = append(out, x)
			}
		}
		if len(out) == 0 {
			delete(h.subs, key)
		} else {
			h.subs[key] = out
		}
		close(c)
	}
}

func writeSSEData(w http.ResponseWriter, fl http.Flusher, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", b); err != nil {
		return err
	}
	fl.Flush()
	return nil
}

func (a *apiDeps) getModelLightsEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	modelID := r.PathValue("id")
	if modelID == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing model id")
		return
	}
	ok, err := a.store.ModelExists(r.Context(), modelID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not load model")
		return
	}
	if !ok {
		writeAPIError(w, http.StatusNotFound, "not_found", "model not found")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	fl, ok := w.(http.Flusher)
	if !ok {
		return
	}

	rc := http.NewResponseController(w)
	extend := func() {
		_ = rc.SetWriteDeadline(time.Now().Add(60 * time.Second))
	}
	extend()

	if _, err := fmt.Fprintf(w, ": dlm lights events\n\n"); err != nil {
		return
	}
	fl.Flush()

	key := modelTopic(modelID)
	last, revCh, unsub := a.rev.subscribe(key)
	defer unsub()

	if err := writeSSEData(w, fl, map[string]uint64{"seq": last}); err != nil {
		return
	}
	extend()

	tick := time.NewTicker(25 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-tick.C:
			extend()
			if _, err := fmt.Fprintf(w, ": keepalive\n\n"); err != nil {
				return
			}
			fl.Flush()
		case seq, ok := <-revCh:
			if !ok {
				return
			}
			extend()
			if err := writeSSEData(w, fl, map[string]uint64{"seq": seq}); err != nil {
				return
			}
		}
	}
}

func (a *apiDeps) getSceneLightsEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	sceneID := r.PathValue("id")
	if sceneID == "" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "missing scene id")
		return
	}
	ok, err := a.store.SceneExists(r.Context(), sceneID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "could not load scene")
		return
	}
	if !ok {
		writeAPIError(w, http.StatusNotFound, "not_found", "scene not found")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	fl, ok := w.(http.Flusher)
	if !ok {
		return
	}

	rc := http.NewResponseController(w)
	extend := func() {
		_ = rc.SetWriteDeadline(time.Now().Add(60 * time.Second))
	}
	extend()

	if _, err := fmt.Fprintf(w, ": dlm scene lights events\n\n"); err != nil {
		return
	}
	fl.Flush()

	key := sceneTopic(sceneID)
	last, revCh, unsub := a.rev.subscribe(key)
	defer unsub()

	if err := writeSSEData(w, fl, map[string]uint64{"seq": last}); err != nil {
		return
	}
	extend()

	tick := time.NewTicker(25 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-tick.C:
			extend()
			if _, err := fmt.Fprintf(w, ": keepalive\n\n"); err != nil {
				return
			}
			fl.Flush()
		case seq, ok := <-revCh:
			if !ok {
				return
			}
			extend()
			if err := writeSSEData(w, fl, map[string]uint64{"seq": seq}); err != nil {
				return
			}
		}
	}
}
