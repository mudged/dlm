// Package reconstruct manages async video-to-model reconstruction jobs.
//
// Jobs are held in memory only; a job lost to a process restart is acceptable
// per the design notes in docs/work-items/WI-06-reconstruction-orchestration.md.
// The confirmed model is the only artefact that is persisted (via the store).
package reconstruct

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"

	"example.com/dlm/backend/internal/cvruntime"
	"example.com/dlm/backend/internal/store"
	"example.com/dlm/backend/internal/wiremodel"
)

// CVRunner is the interface the Manager uses to invoke the CV pipeline.
// The production implementation wraps cvruntime.Run; tests supply a fake.
type CVRunner interface {
	Run(ctx context.Context, spec cvruntime.JobSpec) (cvruntime.Result, error)
}

// RealRunner is the production CVRunner backed by cvruntime.Run.
type RealRunner struct{}

func (RealRunner) Run(ctx context.Context, spec cvruntime.JobSpec) (cvruntime.Result, error) {
	return cvruntime.Run(ctx, spec)
}

// Job status constants.
const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
)

// Coarse progress milestones exposed to clients polling GET /models/capture/{jobId}.
const (
	progressPending  = 0.0
	progressRunning  = 0.1
	progressComplete = 1.0
)

// Resource bounds (Pi defaults; override via Option for tests).
const (
	// maxConcurrentJobs is the default maximum number of CV jobs running at once.
	maxConcurrentJobs = 1
	// defaultMaxRetainedJobs is the default cap on terminal (succeeded/failed) jobs
	// held in memory. Oldest are evicted when the cap is reached.
	defaultMaxRetainedJobs = 20
	// defaultJobTTL is how long a terminal job is retained before the janitor
	// removes it if it was never confirmed or discarded.
	defaultJobTTL = time.Hour
)

// Sentinel errors.
var (
	ErrJobNotFound     = errors.New("reconstruct: job not found")
	ErrJobNotSucceeded = errors.New("reconstruct: job has not succeeded")
	// ErrCapExceeded is returned by Create when the concurrent-job limit is reached.
	ErrCapExceeded = errors.New("reconstruct: concurrent job limit reached")
)

// Job holds the state of one async reconstruction job.
// All fields are safe to read under the Manager mutex; callers receive a pointer
// to the live struct, so reads should be treated as snapshots.
type Job struct {
	ID        string
	Status    string
	Progress  float64
	Result    *cvruntime.Result
	Err       string
	WorkDir   string
	CreatedAt time.Time
	cancel    context.CancelFunc
}

// CreateParams holds optional parameters forwarded to the CV pipeline.
type CreateParams struct {
	Marker    *cvruntime.Marker
	ScaleHint *float64
	DwellMS   int
}

// Option configures a Manager. Use WithNow and WithMaxRetainedJobs in tests.
type Option func(*Manager)

// WithNow replaces the Manager's wall-clock function. Useful for testing TTL eviction.
func WithNow(fn func() time.Time) Option {
	return func(m *Manager) { m.now = fn }
}

// WithMaxRetainedJobs overrides the default cap on retained terminal jobs.
// Primarily for tests that want to exercise eviction without creating 20+ jobs.
func WithMaxRetainedJobs(n int) Option {
	return func(m *Manager) { m.maxRetained = n }
}

// Manager orchestrates async reconstruction jobs.
// It is safe for concurrent use.
type Manager struct {
	runner  CVRunner
	st      *store.Store
	baseDir string // parent dir for per-job work directories

	mu   sync.RWMutex
	jobs map[string]*Job

	// sem is a counting semaphore that limits concurrent CV runs to maxConcurrentJobs.
	// A slot is acquired in Create before the goroutine is spawned and released in
	// runJob once the job reaches a terminal state.
	sem chan struct{}

	now         func() time.Time
	maxRetained int
	jobTTL      time.Duration
}

// New creates a Manager. baseDir is used as the root for per-job work
// directories (e.g. DLM_DATA_DIR/runtime/capture). Stale directories left by
// a previous process are removed on construction.
func New(runner CVRunner, st *store.Store, baseDir string, opts ...Option) *Manager {
	m := &Manager{
		runner:      runner,
		st:          st,
		baseDir:     baseDir,
		jobs:        make(map[string]*Job),
		sem:         make(chan struct{}, maxConcurrentJobs),
		now:         time.Now,
		maxRetained: defaultMaxRetainedJobs,
		jobTTL:      defaultJobTTL,
	}
	for _, o := range opts {
		o(m)
	}
	m.cleanupStaleWorkDirs()
	return m
}

// Create writes each reader to a new per-job work directory, then enqueues the
// CV job asynchronously. At least 2 files are required; fewer returns an error.
// names[i] is used as the on-disk filename for files[i]; an empty element falls
// back to "feed_N". Directory components in names are stripped for safety.
//
// Returns ErrCapExceeded if the concurrent-job limit is already reached.
func (m *Manager) Create(ctx context.Context, files []io.Reader, names []string, params CreateParams) (string, error) {
	if len(files) < 2 {
		return "", fmt.Errorf("at least 2 video files are required")
	}

	// Fail fast if the concurrency cap is already reached.
	select {
	case m.sem <- struct{}{}:
	default:
		return "", ErrCapExceeded
	}

	// semAcquired is true until the spawned goroutine takes over ownership.
	// If anything below fails we release the slot via the defer.
	semOwned := true
	defer func() {
		if semOwned {
			<-m.sem
		}
	}()

	id := uuid.New().String()
	workDir := filepath.Join(m.baseDir, id)
	if err := os.MkdirAll(workDir, 0o750); err != nil {
		return "", fmt.Errorf("reconstruct: create work dir: %w", err)
	}

	var feeds []cvruntime.FeedRef
	for i, r := range files {
		name := fmt.Sprintf("feed_%d", i)
		if i < len(names) && names[i] != "" {
			name = filepath.Base(names[i]) // strip any path traversal
		}
		dst := filepath.Join(workDir, name)
		f, err := os.Create(dst)
		if err != nil {
			_ = os.RemoveAll(workDir)
			return "", fmt.Errorf("reconstruct: create feed file: %w", err)
		}
		if _, err := io.Copy(f, r); err != nil {
			_ = f.Close()
			_ = os.RemoveAll(workDir)
			return "", fmt.Errorf("reconstruct: write feed file: %w", err)
		}
		if err := f.Close(); err != nil {
			_ = os.RemoveAll(workDir)
			return "", fmt.Errorf("reconstruct: close feed file: %w", err)
		}
		feeds = append(feeds, cvruntime.FeedRef{Path: dst})
	}

	jobCtx, cancel := context.WithCancel(context.Background())
	job := &Job{
		ID:        id,
		Status:    StatusPending,
		WorkDir:   workDir,
		CreatedAt: m.now(),
		cancel:    cancel,
	}

	m.mu.Lock()
	m.jobs[id] = job
	m.mu.Unlock()

	// The goroutine now owns the semaphore slot; suppress the defer release.
	semOwned = false
	dwellMS := params.DwellMS
	if dwellMS <= 0 {
		dwellMS = cvruntime.DefaultDwellMS
	}
	go m.runJob(jobCtx, job, cvruntime.JobSpec{
		Feeds:     feeds,
		Marker:    params.Marker,
		ScaleHint: params.ScaleHint,
		DwellMS:   dwellMS,
	})

	return id, nil
}

func (m *Manager) runJob(ctx context.Context, job *Job, spec cvruntime.JobSpec) {
	m.mu.Lock()
	job.Status = StatusRunning
	job.Progress = progressRunning
	m.mu.Unlock()

	result, err := m.runner.Run(ctx, spec)

	m.mu.Lock()
	if err != nil {
		job.Status = StatusFailed
		job.Err = err.Error()
	} else if result.Status == "failed" {
		job.Status = StatusFailed
		if result.Error != nil {
			job.Err = *result.Error
		} else {
			job.Err = "CV pipeline reported failure"
		}
	} else {
		job.Status = StatusSucceeded
		job.Progress = progressComplete
		job.Result = &result
	}
	m.mu.Unlock()

	// Release the concurrency slot so new jobs can be submitted.
	<-m.sem

	// Evict stale jobs now that we have a new terminal job.
	m.RunJanitor()
}

// Get returns the job for the given ID and true, or nil and false if not found.
func (m *Manager) Get(id string) (*Job, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	j, ok := m.jobs[id]
	return j, ok
}

// Confirm validates the candidate lights from a succeeded job and persists the
// result as a named model via the store. Returns:
//   - ErrJobNotFound if the job does not exist.
//   - ErrJobNotSucceeded if the job has not completed successfully.
//   - *wiremodel.ParseError for REQ-005/REQ-007 validation failures.
//   - store.ErrDuplicateName if another model with the same name already exists.
func (m *Manager) Confirm(ctx context.Context, jobID, name string) (store.Summary, error) {
	m.mu.RLock()
	job, ok := m.jobs[jobID]
	m.mu.RUnlock()

	if !ok {
		return store.Summary{}, ErrJobNotFound
	}
	if job.Status != StatusSucceeded || job.Result == nil {
		return store.Summary{}, ErrJobNotSucceeded
	}

	// Surface missing lights from the CV result with a specific, user-actionable
	// error before generic contiguity validation.  A non-empty Missing list means
	// the Python pipeline could not triangulate those light IDs, so the resulting
	// lights slice has gaps that would fail REQ-005 contiguity — name them here
	// rather than emitting an opaque "sequential" failure.
	if len(job.Result.Missing) > 0 {
		return store.Summary{}, wiremodel.MissingIDsError(job.Result.Missing)
	}

	lights := make([]wiremodel.Light, len(job.Result.Lights))
	for i, lp := range job.Result.Lights {
		lights[i] = wiremodel.Light{ID: lp.ID, X: lp.X, Y: lp.Y, Z: lp.Z}
	}

	if err := wiremodel.ValidateLights(lights); err != nil {
		return store.Summary{}, err
	}

	sum, err := m.st.Create(ctx, name, lights)
	if err != nil {
		return store.Summary{}, err
	}

	// Remove the job and its work dir asynchronously after a successful confirm.
	go func() {
		m.mu.Lock()
		delete(m.jobs, jobID)
		m.mu.Unlock()
		_ = os.RemoveAll(job.WorkDir)
	}()

	return *sum, nil
}

// Discard cancels a running job (if applicable) and removes its work directory.
// Returns ErrJobNotFound if the job does not exist.
func (m *Manager) Discard(jobID string) error {
	m.mu.Lock()
	job, ok := m.jobs[jobID]
	if !ok {
		m.mu.Unlock()
		return ErrJobNotFound
	}
	delete(m.jobs, jobID)
	m.mu.Unlock()

	if job.cancel != nil {
		job.cancel()
	}
	_ = os.RemoveAll(job.WorkDir)
	return nil
}

// Shutdown cancels all running jobs.  It is safe to call once; the Manager
// continues to accept new Create calls afterwards (e.g. after a factory reset).
func (m *Manager) Shutdown() {
	m.mu.Lock()
	cancels := make([]context.CancelFunc, 0, len(m.jobs))
	for _, job := range m.jobs {
		if job.cancel != nil {
			cancels = append(cancels, job.cancel)
		}
	}
	m.mu.Unlock()

	for _, cancel := range cancels {
		cancel()
	}
}

// RunJanitor evicts terminal jobs that have exceeded jobTTL or that push the
// total count of retained terminal jobs past maxRetained (oldest evicted first).
// It is called automatically after each job becomes terminal and is also safe
// to call directly in tests.
func (m *Manager) RunJanitor() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.now()

	type entry struct {
		id  string
		job *Job
	}
	var terminal []entry
	for id, job := range m.jobs {
		if job.Status == StatusSucceeded || job.Status == StatusFailed {
			terminal = append(terminal, entry{id, job})
		}
	}

	// Sort oldest-first so cap eviction removes the oldest.
	sort.Slice(terminal, func(i, j int) bool {
		return terminal[i].job.CreatedAt.Before(terminal[j].job.CreatedAt)
	})

	toEvict := make(map[string]*Job)

	// TTL eviction: remove jobs created longer ago than jobTTL.
	for _, e := range terminal {
		if now.Sub(e.job.CreatedAt) > m.jobTTL {
			toEvict[e.id] = e.job
		}
	}

	// Count eviction: if more than maxRetained terminal jobs remain after TTL
	// eviction, remove the oldest until we are within the limit.
	remaining := len(terminal) - len(toEvict)
	for _, e := range terminal {
		if remaining <= m.maxRetained {
			break
		}
		if _, already := toEvict[e.id]; already {
			continue
		}
		toEvict[e.id] = e.job
		remaining--
	}

	for id, job := range toEvict {
		delete(m.jobs, id)
		go os.RemoveAll(job.WorkDir) //nolint:errcheck
	}
}

// cleanupStaleWorkDirs removes directories in baseDir that have no corresponding
// in-memory job. Called once on startup before any jobs are registered.
func (m *Manager) cleanupStaleWorkDirs() {
	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		// baseDir doesn't exist yet — nothing to clean up.
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			_ = os.RemoveAll(filepath.Join(m.baseDir, e.Name()))
		}
	}
}
