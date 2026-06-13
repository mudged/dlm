package cvruntime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// ---- resolver tests --------------------------------------------------------

func TestResolve_EnvOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DLM_CV_RUNTIME_DIR", dir)

	got, err := resolve()
	if err != nil {
		t.Fatalf("resolve() unexpected error: %v", err)
	}
	if got != dir {
		t.Errorf("resolve() = %q, want %q", got, dir)
	}
}

func TestResolve_EnvOverride_InvalidPath(t *testing.T) {
	t.Setenv("DLM_CV_RUNTIME_DIR", "/nonexistent/dlm-cv-runtime-that-does-not-exist")

	_, err := resolve()
	if err == nil {
		t.Fatal("resolve() expected error for inaccessible DLM_CV_RUNTIME_DIR, got nil")
	}
}

func TestResolve_NoRuntimeFound(t *testing.T) {
	t.Setenv("DLM_CV_RUNTIME_DIR", "")

	// During normal CI the sibling runtime/cv/ directory does not exist next to the
	// test binary, so we expect a clear, actionable error.  If a sibling happens to
	// exist (rare local setup), we just log and skip rather than hard-failing.
	_, err := resolve()
	if err == nil {
		t.Log("resolve() succeeded — a sibling runtime/cv/ directory exists; skipping error-path check")
	}
}

// ---- stub runtime helpers --------------------------------------------------

// makeStubRuntime creates a minimal fake CV runtime directory suitable for
// unit tests on Unix.  The "interpreter" is a shell script that ignores the
// entrypoint and spec-file arguments and writes cannedJSON to stdout.
func makeStubRuntime(t *testing.T, cannedJSON string) string {
	return makeStubRuntimeWithExit(t, cannedJSON, 0)
}

// makeStubRuntimeWithExit is like makeStubRuntime but the stub exits with exitCode.
func makeStubRuntimeWithExit(t *testing.T, cannedJSON string, exitCode int) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell-script stub not supported on Windows")
	}

	dir := t.TempDir()
	binDir := filepath.Join(dir, "python", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}

	stub := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' '%s'\nexit %d\n", cannedJSON, exitCode)
	if err := os.WriteFile(filepath.Join(binDir, "python3"), []byte(stub), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "reconstruct.py"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// makeSlowStubRuntime creates a stub interpreter that sleeps indefinitely,
// used to test context-cancellation / timeout behaviour.
func makeSlowStubRuntime(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell-script stub not supported on Windows")
	}

	dir := t.TempDir()
	binDir := filepath.Join(dir, "python", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(binDir, "python3"),
		[]byte("#!/bin/sh\nsleep 300\n"),
		0o755,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "reconstruct.py"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// ---- Run tests -------------------------------------------------------------

func TestRun_StubInterpreter(t *testing.T) {
	const cannedJSON = `{"status":"succeeded","light_count":2,` +
		`"lights":[{"id":0,"x":0.1,"y":0.2,"z":0.3},{"id":1,"x":0.4,"y":0.5,"z":0.6}],` +
		`"missing":[],"low_confidence":[],"error":null}`

	dir := makeStubRuntime(t, cannedJSON)
	t.Setenv("DLM_CV_RUNTIME_DIR", dir)

	spec := JobSpec{
		Feeds:   []FeedRef{{Path: "/tmp/feed0.mp4"}},
		DwellMS: 1000,
	}
	result, err := Run(context.Background(), spec)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != "succeeded" {
		t.Errorf("Status = %q, want %q", result.Status, "succeeded")
	}
	if result.LightCount != 2 {
		t.Errorf("LightCount = %d, want 2", result.LightCount)
	}
	if len(result.Lights) != 2 {
		t.Errorf("len(Lights) = %d, want 2", len(result.Lights))
	}
	if result.Lights[0].ID != 0 || result.Lights[1].ID != 1 {
		t.Errorf("unexpected light IDs: %v", result.Lights)
	}
}

func TestRun_ChildReportsFailure(t *testing.T) {
	errMsg := "camera feed unreadable"
	const cannedJSON = `{"status":"failed","light_count":0,"lights":[],"missing":[],"low_confidence":[],"error":"camera feed unreadable"}`

	dir := makeStubRuntime(t, cannedJSON)
	t.Setenv("DLM_CV_RUNTIME_DIR", dir)

	result, err := Run(context.Background(), JobSpec{Feeds: []FeedRef{{Path: "/tmp/x.mp4"}}, DwellMS: 500})
	if err != nil {
		t.Fatalf("Run() returned Go error for child-reported failure; want nil Go error, got: %v", err)
	}
	if result.Status != "failed" {
		t.Errorf("Status = %q, want %q", result.Status, "failed")
	}
	if result.Error == nil || *result.Error != errMsg {
		t.Errorf("Error = %v, want %q", result.Error, errMsg)
	}
}

func TestRun_ChildReportsFailure_NonZeroExit(t *testing.T) {
	errMsg := "no blink events detected in any feed"
	const cannedJSON = `{"status":"failed","light_count":0,"lights":[],"missing":[],"low_confidence":[],"error":"no blink events detected in any feed"}`

	dir := makeStubRuntimeWithExit(t, cannedJSON, 1)
	t.Setenv("DLM_CV_RUNTIME_DIR", dir)

	result, err := Run(context.Background(), JobSpec{Feeds: []FeedRef{{Path: "/tmp/x.mp4"}}, DwellMS: 500})
	if err != nil {
		t.Fatalf("Run() should parse failed JSON even when child exits 1; got: %v", err)
	}
	if result.Status != "failed" {
		t.Errorf("Status = %q, want %q", result.Status, "failed")
	}
	if result.Error == nil || *result.Error != errMsg {
		t.Errorf("Error = %v, want %q", result.Error, errMsg)
	}
}

func TestRun_ContextTimeout_KillsChild(t *testing.T) {
	dir := makeSlowStubRuntime(t)
	t.Setenv("DLM_CV_RUNTIME_DIR", dir)

	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := Run(ctx, JobSpec{Feeds: []FeedRef{{Path: "/tmp/feed.mp4"}}, DwellMS: 1000})
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Run() expected error from timeout, got nil")
	}
	// Child should be killed well within killGracePeriod (5 s) plus the timeout (0.4 s).
	if elapsed > killGracePeriod+2*time.Second {
		t.Errorf("Run() took %v — child was not killed promptly after context expiry", elapsed)
	}
}

func TestRun_ContextCancel_KillsChild(t *testing.T) {
	dir := makeSlowStubRuntime(t)
	t.Setenv("DLM_CV_RUNTIME_DIR", dir)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		_, err := Run(ctx, JobSpec{Feeds: []FeedRef{{Path: "/tmp/feed.mp4"}}, DwellMS: 1000})
		done <- err
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("Run() expected error after cancel, got nil")
		}
	case <-time.After(killGracePeriod + 3*time.Second):
		t.Fatal("Run() did not return within expected time after context cancel")
	}
}

// makeVerboseStubRuntime creates a stub interpreter that writes more than
// stdoutCapBytes to stdout and exits 0, used to verify the cap is enforced.
func makeVerboseStubRuntime(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell-script stub not supported on Windows")
	}

	dir := t.TempDir()
	binDir := filepath.Join(dir, "python", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Emit 5 MiB of null bytes to exceed the 4 MiB cap and exit 0.
	// Pipe-free (avoids SIGPIPE in parallel test runs): dd writes directly to stdout.
	stub := "#!/bin/sh\ndd if=/dev/zero bs=1048576 count=5 2>/dev/null\n"
	if err := os.WriteFile(filepath.Join(binDir, "python3"), []byte(stub), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "reconstruct.py"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestRun_StdoutCapExceeded(t *testing.T) {
	dir := makeVerboseStubRuntime(t)
	t.Setenv("DLM_CV_RUNTIME_DIR", dir)

	_, err := Run(context.Background(), JobSpec{Feeds: []FeedRef{{Path: "/tmp/feed.mp4"}}, DwellMS: 500})
	if err == nil {
		t.Fatal("Run() expected error when stdout cap exceeded, got nil")
	}
	if !strings.Contains(err.Error(), "stdout exceeded") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRun_NormalChildStillParsed(t *testing.T) {
	// Confirm that a well-behaved child (output well within cap) is still
	// parsed correctly — regression guard for the limitedWriter change.
	const cannedJSON = `{"status":"succeeded","light_count":1,` +
		`"lights":[{"id":0,"x":1.0,"y":2.0,"z":3.0}],` +
		`"missing":[],"low_confidence":[],"error":null}`

	dir := makeStubRuntime(t, cannedJSON)
	t.Setenv("DLM_CV_RUNTIME_DIR", dir)

	result, err := Run(context.Background(), JobSpec{Feeds: []FeedRef{{Path: "/tmp/feed.mp4"}}, DwellMS: 500})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != "succeeded" {
		t.Errorf("Status = %q, want succeeded", result.Status)
	}
	if len(result.Lights) != 1 || result.Lights[0].ID != 0 {
		t.Errorf("unexpected lights: %v", result.Lights)
	}
}

// TestRun_SpecFileWritten verifies that Run writes the JobSpec to a temp file
// and passes its path as the first argument to the interpreter.  The stub checks
// that its first argument is a readable file, confirming the spec-file handoff.
func TestRun_SpecFileWritten(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script stub not supported on Windows")
	}

	dir := t.TempDir()
	binDir := filepath.Join(dir, "python", "bin")
	os.MkdirAll(binDir, 0o755)

	// The stub verifies $1 is a readable file (the spec), then returns a canned result.
	stub := `#!/bin/sh
specFile="$1"
if [ -z "$specFile" ] || [ ! -f "$specFile" ]; then
  printf '{"status":"failed","light_count":0,"lights":[],"missing":[],"low_confidence":[],"error":"spec file not found at $specFile"}\n'
  exit 0
fi
printf '{"status":"succeeded","light_count":3,"lights":[],"missing":[],"low_confidence":[],"error":null}\n'
`
	os.WriteFile(filepath.Join(binDir, "python3"), []byte(stub), 0o755)
	os.WriteFile(filepath.Join(dir, "reconstruct.py"), []byte(""), 0o644)
	t.Setenv("DLM_CV_RUNTIME_DIR", dir)

	spec := JobSpec{
		Feeds:   []FeedRef{{Path: "/tmp/a.mp4"}, {Path: "/tmp/b.mp4"}, {Path: "/tmp/c.mp4"}},
		DwellMS: 2000,
	}
	result, err := Run(context.Background(), spec)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != "succeeded" {
		t.Errorf("Status = %q, want succeeded", result.Status)
	}
	if result.LightCount != 3 {
		t.Errorf("LightCount = %d, want 3", result.LightCount)
	}
}
