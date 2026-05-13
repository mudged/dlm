package httpapi

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"example.com/dlm/backend/internal/wiremodel"
)

// captureHandler is a minimal slog.Handler that records every record in memory
// so tests can assert on level + attributes (REQ-041 BR 6 fan-out logging).
type captureHandler struct {
	mu      sync.Mutex
	records []slog.Record
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records = append(h.records, r.Clone())
	return nil
}
func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(string) slog.Handler      { return h }

func (h *captureHandler) snapshot() []slog.Record {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]slog.Record, len(h.records))
	copy(out, h.records)
	return out
}

// drainSubscriber consumes payloads from one revision hub subscriber for assertions.
func drainSubscriber(ch <-chan LightsSSEPayload, max int, timeout time.Duration) []LightsSSEPayload {
	out := make([]LightsSSEPayload, 0, max)
	deadline := time.After(timeout)
	for len(out) < max {
		select {
		case msg, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, msg)
		case <-deadline:
			return out
		}
	}
	return out
}

func TestRevisionHub_NotifyModelLightsChanged_logsWhenSceneLookupFails(t *testing.T) {
	ctx := context.Background()
	st := testStore(t)
	sum, err := st.Create(ctx, "fanout-fail", []wiremodel.Light{{ID: 0, X: 0, Y: 0, Z: 0}})
	if err != nil {
		t.Fatal(err)
	}

	cap := &captureHandler{}
	log := slog.New(cap)
	hub := NewRevisionHubWithLogger(log)

	// Subscribe BEFORE breaking the store so we can assert the model topic still ticks.
	lastSeq, ch, unsub := hub.subscribe(modelTopic(sum.ID))
	defer unsub()
	if lastSeq != 0 {
		t.Fatalf("lastSeq before any emit = %d want 0", lastSeq)
	}

	if err := st.Close(); err != nil {
		t.Fatalf("close store to force scene lookup error: %v", err)
	}

	deltas := []LightsSSEDelta{{LightID: 0, On: true, Color: "#ff0000", BrightnessPct: 100}}
	hub.NotifyModelLightsChanged(ctx, st, sum.ID, deltas)

	msgs := drainSubscriber(ch, 1, time.Second)
	if len(msgs) != 1 {
		t.Fatalf("model topic did not receive delta after fan-out failure: got %d msgs", len(msgs))
	}
	if got := msgs[0].Seq; got != 1 {
		t.Fatalf("model topic seq = %d want 1", got)
	}
	if len(msgs[0].Deltas) != 1 || msgs[0].Deltas[0].LightID != 0 {
		t.Fatalf("unexpected model delta payload: %+v", msgs[0])
	}

	recs := cap.snapshot()
	if len(recs) == 0 {
		t.Fatalf("expected at least one slog record after fan-out failure")
	}
	var found bool
	for _, r := range recs {
		if r.Level < slog.LevelWarn {
			continue
		}
		if !strings.Contains(r.Message, "scene fan-out") {
			continue
		}
		var hasModel, hasErr bool
		r.Attrs(func(a slog.Attr) bool {
			switch a.Key {
			case "model_id":
				if a.Value.String() == sum.ID {
					hasModel = true
				}
			case "err":
				if a.Value.String() != "" {
					hasErr = true
				}
			}
			return true
		})
		if hasModel && hasErr {
			found = true
			break
		}
	}
	if !found {
		var buf bytes.Buffer
		for _, r := range recs {
			buf.WriteString(r.Level.String())
			buf.WriteByte(' ')
			buf.WriteString(r.Message)
			r.Attrs(func(a slog.Attr) bool {
				buf.WriteByte(' ')
				buf.WriteString(a.Key)
				buf.WriteByte('=')
				buf.WriteString(a.Value.String())
				return true
			})
			buf.WriteByte('\n')
		}
		t.Fatalf("expected WARN log with scene fan-out + model_id=%s + err; got:\n%s", sum.ID, buf.String())
	}
}
