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

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return Result{}, fmt.Errorf("cvruntime: child exited with error: %w\nstderr: %s", err, stderr.String())
		}
		return Result{}, fmt.Errorf("cvruntime: child exited with error: %w", err)
	}

	var result Result
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return Result{}, fmt.Errorf(
			"cvruntime: failed to parse child JSON output: %w\nraw stdout: %s",
			err, stdout.String(),
		)
	}
	return result, nil
}
