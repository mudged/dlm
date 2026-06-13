package cvruntime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// killGracePeriod is the time the child gets to respond to SIGTERM before
// exec.Cmd.WaitDelay escalates to SIGKILL.
const killGracePeriod = 5 * time.Second

// stdoutCapBytes is the maximum number of bytes buffered from the child's
// stdout before the job is treated as failed. Protects against runaway output.
const stdoutCapBytes = 4 << 20 // 4 MiB

// stderrTailBytes is how many trailing bytes of stderr are kept for diagnostics.
const stderrTailBytes = 64 << 10 // 64 KiB

// limitedWriter caps total bytes written to its internal buffer at max.
// Writes beyond the cap are silently discarded (len(p), nil is still returned
// so the caller — exec internals — does not see a write error mid-run).
// Truncated is set to true whenever bytes are dropped.
type limitedWriter struct {
	buf       bytes.Buffer
	max       int
	Truncated bool
}

func (lw *limitedWriter) Write(p []byte) (int, error) {
	remaining := lw.max - lw.buf.Len()
	if remaining <= 0 {
		lw.Truncated = true
		return len(p), nil
	}
	if len(p) > remaining {
		lw.Truncated = true
		p = p[:remaining]
	}
	return lw.buf.Write(p)
}

// tailBuffer keeps only the last max bytes of all written data.
type tailBuffer struct {
	data []byte
	max  int
}

func (tb *tailBuffer) Write(p []byte) (int, error) {
	tb.data = append(tb.data, p...)
	if len(tb.data) > tb.max {
		tb.data = tb.data[len(tb.data)-tb.max:]
	}
	return len(p), nil
}

func (tb *tailBuffer) Len() int      { return len(tb.data) }
func (tb *tailBuffer) String() string { return string(tb.data) }

// Run resolves the bundled CV runtime, launches the reconstruct entrypoint as a
// supervised child process, and returns the parsed Result.
//
// Context cancellation or timeout causes Run to send SIGTERM to the child; if
// the child does not exit within killGracePeriod it is SIGKILLed.
func Run(ctx context.Context, spec JobSpec) (Result, error) {
	runtimeDir, err := resolve()
	if err != nil {
		return Result{}, err
	}
	return runWithDir(ctx, spec, runtimeDir)
}

// runWithDir is the testable core of Run; callers supply the resolved runtimeDir.
func runWithDir(ctx context.Context, spec JobSpec, runtimeDir string) (Result, error) {
	specData, err := json.Marshal(spec)
	if err != nil {
		return Result{}, fmt.Errorf("cvruntime: marshal spec: %w", err)
	}

	specFile, err := os.CreateTemp("", "dlm-cvspec-*.json")
	if err != nil {
		return Result{}, fmt.Errorf("cvruntime: create spec temp file: %w", err)
	}
	specPath := specFile.Name()
	defer os.Remove(specPath)

	if _, err := specFile.Write(specData); err != nil {
		specFile.Close()
		return Result{}, fmt.Errorf("cvruntime: write spec file: %w", err)
	}
	if err := specFile.Close(); err != nil {
		return Result{}, fmt.Errorf("cvruntime: close spec file: %w", err)
	}

	interp := interpreterPath(runtimeDir)
	entry := entrypointPath(runtimeDir)

	cmd := exec.CommandContext(ctx, interp, entry, specPath)

	// On context cancellation: send SIGTERM first; WaitDelay escalates to SIGKILL.
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return gracefulSignal(cmd.Process)
	}
	cmd.WaitDelay = killGracePeriod

	stdout := &limitedWriter{max: stdoutCapBytes}
	stderr := &tailBuffer{max: stderrTailBytes}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		// Logical failures emit status:"failed" JSON on stdout and exit non-zero.
		// Parse that payload so callers can surface result.Error instead of a
		// generic "child exited" message.
		if stdout.buf.Len() > 0 {
			var result Result
			if jsonErr := json.Unmarshal(stdout.buf.Bytes(), &result); jsonErr == nil &&
				result.Status == "failed" {
				return result, nil
			}
		}
		if stderr.Len() > 0 {
			return Result{}, fmt.Errorf("cvruntime: child exited with error: %w\nstderr: %s", err, stderr.String())
		}
		return Result{}, fmt.Errorf("cvruntime: child exited with error: %w", err)
	}

	if stdout.Truncated {
		return Result{}, fmt.Errorf(
			"cvruntime: child stdout exceeded %d-byte cap; possible runaway output",
			stdoutCapBytes,
		)
	}

	var result Result
	if err := json.Unmarshal(stdout.buf.Bytes(), &result); err != nil {
		return Result{}, fmt.Errorf(
			"cvruntime: failed to parse child JSON output: %w\nraw stdout: %s",
			err, stdout.buf.String(),
		)
	}
	return result, nil
}
