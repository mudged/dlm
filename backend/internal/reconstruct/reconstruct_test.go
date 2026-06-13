package reconstruct_test

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"example.com/dlm/backend/internal/cvruntime"
	"example.com/dlm/backend/internal/lightstate"
	"example.com/dlm/backend/internal/reconstruct"
	"example.com/dlm/backend/internal/store"
	"example.com/dlm/backend/internal/wiremodel"
)

// fakeRunner is an injectable CVRunner for tests.
type fakeRunner struct {
	result cvruntime.Result
	err    error
}

func (f *fakeRunner) Run(_ context.Context, _ cvruntime.JobSpec) (cvruntime.Result, error) {
	return f.result, f.err
}

func goodResult(n int) cvruntime.Result {
	lights := make([]cvruntime.LightPoint, n)
	for i := range lights {
		lights[i] = cvruntime.LightPoint{ID: i, X: float64(i), Y: 0, Z: 0}
	}
	return cvruntime.Result{
		Status:     "succeeded",
		LightCount: n,
		Lights:     lights,
	}
}

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	p := filepath.Join(t.TempDir(), "test.db")
	s, err := store.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	ls := lightstate.New()
	s.SetLightState(ls)
	if err := s.LoadLightStateFromDB(context.Background()); err != nil {
		t.Fatal(err)
	}
	return s
}

func readers(strs ...string) []io.Reader {
	var rs []io.Reader
	for _, s := range strs {
		rs = append(rs, strings.NewReader(s))
	}
	return rs
}

// waitFor polls until the job reaches the desired status or times out.
func waitFor(t *testing.T, m *reconstruct.Manager, id, wantStatus string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		j, ok := m.Get(id)
		if !ok {
			t.Fatalf("job %s disappeared while waiting for %s", id, wantStatus)
		}
		if j.Status == wantStatus {
			return
		}
		// Don't wait if the job failed when we wanted succeeded (or vice-versa).
		if j.Status == reconstruct.StatusFailed && wantStatus == reconstruct.StatusSucceeded {
			t.Fatalf("job failed unexpectedly: %s", j.Err)
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("job %s did not reach %s within timeout", id, wantStatus)
}

// waitTerminal polls until the job reaches any terminal state.
func waitTerminal(t *testing.T, m *reconstruct.Manager, id string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		j, ok := m.Get(id)
		if !ok {
			t.Fatalf("job %s disappeared", id)
		}
		if j.Status == reconstruct.StatusSucceeded || j.Status == reconstruct.StatusFailed {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("job %s did not reach a terminal state within timeout", id)
}

func TestManager_happyPath(t *testing.T) {
	st := newTestStore(t)
	runner := &fakeRunner{result: goodResult(3)}
	m := reconstruct.New(runner, st, t.TempDir())

	id, err := m.Create(context.Background(), readers("v1", "v2"), []string{"a.mp4", "b.mp4"}, reconstruct.CreateParams{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty job id")
	}

	waitFor(t, m, id, reconstruct.StatusSucceeded)

	job, ok := m.Get(id)
	if !ok {
		t.Fatal("job not found after succeeding")
	}
	if job.Status != reconstruct.StatusSucceeded {
		t.Fatalf("status = %q, want succeeded", job.Status)
	}
	if job.Progress != 1.0 {
		t.Fatalf("progress = %v, want 1.0", job.Progress)
	}

	sum, err := m.Confirm(context.Background(), id, "my-model")
	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if sum.Name != "my-model" {
		t.Fatalf("summary name = %q", sum.Name)
	}
	if sum.LightCount != 3 {
		t.Fatalf("light count = %d, want 3", sum.LightCount)
	}

	// Verify the model is in the store.
	all, err := st.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, s := range all {
		if s.ID == sum.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("model not found in store after confirm")
	}
}

func TestManager_statusTransitions(t *testing.T) {
	st := newTestStore(t)
	runner := &fakeRunner{result: goodResult(2)}
	m := reconstruct.New(runner, st, t.TempDir())

	id, err := m.Create(context.Background(), readers("v1", "v2"), nil, reconstruct.CreateParams{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Initially pending or already running (goroutine may have started).
	job, ok := m.Get(id)
	if !ok {
		t.Fatal("job not found immediately after Create")
	}
	if job.Status != reconstruct.StatusPending && job.Status != reconstruct.StatusRunning {
		t.Fatalf("initial status = %q, want pending or running", job.Status)
	}

	waitFor(t, m, id, reconstruct.StatusSucceeded)
}

func TestManager_fewerThan2Files_rejected(t *testing.T) {
	st := newTestStore(t)
	m := reconstruct.New(&fakeRunner{}, st, t.TempDir())

	_, err := m.Create(context.Background(), readers("single"), []string{"a.mp4"}, reconstruct.CreateParams{})
	if err == nil {
		t.Fatal("expected error for fewer than 2 files")
	}
}

func TestManager_confirmNonSucceeded_error(t *testing.T) {
	st := newTestStore(t)
	runner := &fakeRunner{err: errors.New("cv pipeline blew up")}
	m := reconstruct.New(runner, st, t.TempDir())

	id, err := m.Create(context.Background(), readers("v1", "v2"), nil, reconstruct.CreateParams{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	waitTerminal(t, m, id)

	job, _ := m.Get(id)
	if job.Status != reconstruct.StatusFailed {
		t.Fatalf("status = %q, want failed", job.Status)
	}

	_, err = m.Confirm(context.Background(), id, "name")
	if err == nil {
		t.Fatal("expected error confirming non-succeeded job")
	}
	if !errors.Is(err, reconstruct.ErrJobNotSucceeded) {
		t.Fatalf("expected ErrJobNotSucceeded, got: %v", err)
	}
}

func TestManager_confirmInvalidCandidate_validationError(t *testing.T) {
	st := newTestStore(t)
	// Non-sequential IDs (gap between 0 and 5) will fail ValidateLights.
	bad := cvruntime.Result{
		Status: "succeeded",
		Lights: []cvruntime.LightPoint{
			{ID: 0}, {ID: 5},
		},
	}
	runner := &fakeRunner{result: bad}
	m := reconstruct.New(runner, st, t.TempDir())

	id, err := m.Create(context.Background(), readers("v1", "v2"), nil, reconstruct.CreateParams{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	waitFor(t, m, id, reconstruct.StatusSucceeded)

	_, err = m.Confirm(context.Background(), id, "bad-model")
	var pe *wiremodel.ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *wiremodel.ParseError, got %v (%T)", err, err)
	}
}

func TestManager_confirmUnknownJob_error(t *testing.T) {
	st := newTestStore(t)
	m := reconstruct.New(&fakeRunner{}, st, t.TempDir())

	_, err := m.Confirm(context.Background(), "no-such-id", "name")
	if !errors.Is(err, reconstruct.ErrJobNotFound) {
		t.Fatalf("expected ErrJobNotFound, got: %v", err)
	}
}

func TestManager_discard_removesWorkDir(t *testing.T) {
	st := newTestStore(t)
	base := t.TempDir()
	runner := &fakeRunner{result: goodResult(2)}
	m := reconstruct.New(runner, st, base)

	id, err := m.Create(context.Background(), readers("v1", "v2"), nil, reconstruct.CreateParams{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Grab the work dir before discarding.
	job, ok := m.Get(id)
	if !ok {
		t.Fatal("job not found")
	}
	workDir := job.WorkDir

	if err := m.Discard(id); err != nil {
		t.Fatalf("Discard: %v", err)
	}

	if _, err := os.Stat(workDir); !os.IsNotExist(err) {
		t.Fatal("work dir still exists after discard")
	}
	_, ok = m.Get(id)
	if ok {
		t.Fatal("job still in manager after discard")
	}
}

func TestManager_discardUnknown_error(t *testing.T) {
	st := newTestStore(t)
	m := reconstruct.New(&fakeRunner{}, st, t.TempDir())

	if err := m.Discard("no-such-id"); !errors.Is(err, reconstruct.ErrJobNotFound) {
		t.Fatalf("expected ErrJobNotFound, got: %v", err)
	}
}
