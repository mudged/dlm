package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"example.com/dlm/backend/internal/store"
)

// LightsSSEDelta is one light's full output triple after a successful commit (REQ-041).
// ModelID is omitted on model-scoped SSE; required on scene-scoped SSE.
type LightsSSEDelta struct {
	ModelID       string  `json:"model_id,omitempty"`
	LightID       int     `json:"light_id"`
	On            bool    `json:"on"`
	Color         string  `json:"color"`
	BrightnessPct float64 `json:"brightness_pct"`
}

// LightsSSEPayload is the JSON object on each SSE data line (REQ-041).
type LightsSSEPayload struct {
	Seq    uint64           `json:"seq"`
	Deltas []LightsSSEDelta `json:"deltas"`
}

// subscriber is one SSE client's receive end. ch carries payloads; done is closed
// (once) to signal the SSE handler goroutine to exit when the subscriber is too slow.
type subscriber struct {
	ch   chan LightsSSEPayload
	done chan struct{}
	once sync.Once
}

// disconnect closes done exactly once so the SSE handler goroutine unblocks and exits.
func (s *subscriber) disconnect() {
	s.once.Do(func() { close(s.done) })
}

// RevisionHub broadcasts monotonic revision numbers and per-commit deltas for model and
// scene light-state topics (REQ-029, REQ-041). Safe for concurrent use.
type RevisionHub struct {
	mu   sync.Mutex
	seq  map[string]uint64
	subs map[string][]*subscriber
	log  *slog.Logger
}

// NewRevisionHub constructs an empty hub. Fan-out lookup failures are reported
// via slog.Default(); use NewRevisionHubWithLogger to inject a different logger
// (e.g. the JSON handler wired in router.go).
func NewRevisionHub() *RevisionHub {
	return NewRevisionHubWithLogger(nil)
}

// NewRevisionHubWithLogger constructs an empty hub using the supplied logger
// for REQ-041 BR 6 fan-out warnings. A nil logger falls back to slog.Default().
func NewRevisionHubWithLogger(log *slog.Logger) *RevisionHub {
	if log == nil {
		log = slog.Default()
	}
	return &RevisionHub{
		seq:  make(map[string]uint64),
		subs: make(map[string][]*subscriber),
		log:  log,
	}
}

func (h *RevisionHub) logger() *slog.Logger {
	if h == nil || h.log == nil {
		return slog.Default()
	}
	return h.log
}

func modelTopic(modelID string) string { return "m:" + modelID }
func sceneTopic(sceneID string) string { return "s:" + sceneID }

func sceneStateToDelta(ps store.ScenePatchedState) LightsSSEDelta {
	return LightsSSEDelta{
		ModelID:       ps.ModelID,
		LightID:       ps.ID,
		On:            ps.On,
		Color:         ps.Color,
		BrightnessPct: ps.BrightnessPct,
	}
}

func lightDTOToDelta(d store.LightStateDTO) LightsSSEDelta {
	return LightsSSEDelta{
		LightID:       d.ID,
		On:            d.On,
		Color:         d.Color,
		BrightnessPct: d.BrightnessPct,
	}
}

func cloneDeltasWithModelID(modelID string, src []LightsSSEDelta) []LightsSSEDelta {
	out := make([]LightsSSEDelta, len(src))
	for i := range src {
		out[i] = src[i]
		out[i].ModelID = modelID
	}
	return out
}

func (h *RevisionHub) emit(key string, deltas []LightsSSEDelta) {
	if h == nil || key == "" {
		return
	}
	if deltas == nil {
		deltas = []LightsSSEDelta{}
	}
	h.mu.Lock()
	h.seq[key]++
	v := h.seq[key]
	msg := LightsSSEPayload{Seq: v, Deltas: deltas}
	subs := append([]*subscriber(nil), h.subs[key]...)
	h.mu.Unlock()

	// Do not hold mu while sending: unsubscribe needs the lock.
	// Non-blocking send: a full or concurrently-closed channel goes to default.
	// A subscriber that would block is disconnected so it cannot stall API handlers.
	for _, sub := range subs {
		if !emitSendPayload(sub.ch, msg) {
			sub.disconnect()
			h.logger().Warn("revision_hub: slow subscriber disconnected; client should reconnect",
				"key", key,
				"seq", v,
			)
		}
	}
}

// emitSendPayload attempts a non-blocking send. Returns true if the message was
// enqueued, false if the channel buffer was full (or the channel was already closed).
// Recovers from the panic caused by sending on a closed channel (concurrent unsubscribe).
func emitSendPayload(ch chan LightsSSEPayload, msg LightsSSEPayload) (sent bool) {
	defer func() { recover() }()
	select {
	case ch <- msg:
		return true
	default:
		return false
	}
}

// NotifyModelLightsChanged bumps the model topic and every scene that contains this model.
// Deltas must use empty ModelID (implicit on GET …/models/{id}/lights/events).
func (h *RevisionHub) NotifyModelLightsChanged(ctx context.Context, st *store.Store, modelID string, deltas []LightsSSEDelta) {
	if h == nil || st == nil || modelID == "" || len(deltas) == 0 {
		return
	}
	h.emit(modelTopic(modelID), deltas)
	scenes, err := st.ListSceneIDsForModel(ctx, modelID)
	if err != nil {
		// REQ-041 BR 6: scene subscribers will not see this delta until the
		// next change. Surface so operators can correlate stuck SSE streams.
		h.logger().Warn("revision_hub: scene fan-out lookup failed",
			"model_id", modelID,
			"err", err,
		)
		return
	}
	sceneDeltas := cloneDeltasWithModelID(modelID, deltas)
	for _, sid := range scenes {
		h.emit(sceneTopic(sid), sceneDeltas)
	}
}

// NotifyAfterSceneLightPatch notifies the patched scene and propagates model / other scene topics (REQ-041).
func (h *RevisionHub) NotifyAfterSceneLightPatch(ctx context.Context, st *store.Store, sceneID string, states []store.ScenePatchedState) {
	if h == nil || st == nil || sceneID == "" || len(states) == 0 {
		return
	}
	sceneDeltas := make([]LightsSSEDelta, 0, len(states))
	for _, ps := range states {
		sceneDeltas = append(sceneDeltas, sceneStateToDelta(ps))
	}
	h.emit(sceneTopic(sceneID), sceneDeltas)

	byModel := make(map[string][]LightsSSEDelta)
	for _, ps := range states {
		if ps.ModelID == "" {
			continue
		}
		byModel[ps.ModelID] = append(byModel[ps.ModelID], LightsSSEDelta{
			LightID:       ps.ID,
			On:            ps.On,
			Color:         ps.Color,
			BrightnessPct: ps.BrightnessPct,
		})
	}
	for mid, md := range byModel {
		if len(md) == 0 {
			continue
		}
		h.emit(modelTopic(mid), md)
		others, err := st.ListSceneIDsForModel(ctx, mid)
		if err != nil {
			// REQ-041 BR 6: log so the missed cross-scene fan-out is visible.
			h.logger().Warn("revision_hub: scene fan-out lookup failed",
				"model_id", mid,
				"err", err,
			)
			continue
		}
		for _, sid := range others {
			if sid == sceneID {
				continue
			}
			h.emit(sceneTopic(sid), cloneDeltasWithModelID(mid, md))
		}
	}
}

// subscribe registers a new SSE subscriber for the given topic key.
// The caller must invoke unsub (exactly once, typically via defer) to remove the
// subscriber and release its channel. The returned done channel is closed if emit
// detects the subscriber is too slow; the SSE handler should exit when done fires.
func (h *RevisionHub) subscribe(key string) (lastSeq uint64, ch <-chan LightsSSEPayload, done <-chan struct{}, unsub func()) {
	sub := &subscriber{
		ch:   make(chan LightsSSEPayload, 1024),
		done: make(chan struct{}),
	}
	h.mu.Lock()
	lastSeq = h.seq[key]
	h.subs[key] = append(h.subs[key], sub)
	h.mu.Unlock()
	return lastSeq, sub.ch, sub.done, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		lst := h.subs[key]
		out := lst[:0]
		for _, x := range lst {
			if x != sub {
				out = append(out, x)
			}
		}
		if len(out) == 0 {
			delete(h.subs, key)
		} else {
			h.subs[key] = out
		}
		close(sub.ch)
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
	last, revCh, done, unsub := a.rev.subscribe(key)
	defer unsub()

	if err := writeSSEData(w, fl, LightsSSEPayload{Seq: last, Deltas: []LightsSSEDelta{}}); err != nil {
		return
	}
	extend()

	tick := time.NewTicker(25 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-done:
			return
		case <-tick.C:
			extend()
			if _, err := fmt.Fprintf(w, ": keepalive\n\n"); err != nil {
				return
			}
			fl.Flush()
		case msg, ok := <-revCh:
			if !ok {
				return
			}
			extend()
			if err := writeSSEData(w, fl, msg); err != nil {
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
	last, revCh, done, unsub := a.rev.subscribe(key)
	defer unsub()

	if err := writeSSEData(w, fl, LightsSSEPayload{Seq: last, Deltas: []LightsSSEDelta{}}); err != nil {
		return
	}
	extend()

	tick := time.NewTicker(25 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-done:
			return
		case <-tick.C:
			extend()
			if _, err := fmt.Fprintf(w, ": keepalive\n\n"); err != nil {
				return
			}
			fl.Flush()
		case msg, ok := <-revCh:
			if !ok {
				return
			}
			extend()
			if err := writeSSEData(w, fl, msg); err != nil {
				return
			}
		}
	}
}
