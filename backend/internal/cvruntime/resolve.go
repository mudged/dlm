package cvruntime

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// resolve returns the absolute path to the CV runtime directory.
//
// Resolution order:
//  1. DLM_CV_RUNTIME_DIR environment variable (absolute path) — used in dev and tests.
//  2. <dir-of-executable>/runtime/cv/  — mechanism B: sibling asset in the release archive.
//
// Returns a descriptive, actionable error if no usable directory is found so that
// reconstruction fails with a helpful message rather than a panic.
func resolve() (string, error) {
	// 1. Explicit env override; useful in dev and always used in package tests.
	if dir := os.Getenv("DLM_CV_RUNTIME_DIR"); dir != "" {
		if _, err := os.Stat(dir); err != nil {
			return "", fmt.Errorf(
				"cvruntime: DLM_CV_RUNTIME_DIR=%q is set but the directory is not accessible: %w",
				dir, err,
			)
		}
		return dir, nil
	}

	// 2. Sibling to the running executable (mechanism B).
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("cvruntime: cannot determine executable path: %w", err)
	}
	// EvalSymlinks gives the real binary path when running under a symlink (e.g. go run).
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("cvruntime: cannot resolve executable symlinks: %w", err)
	}
	candidate := filepath.Join(filepath.Dir(exe), "runtime", "cv")
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}

	return "", fmt.Errorf(
		"cvruntime: no CV runtime bundle found for %s/%s; "+
			"set DLM_CV_RUNTIME_DIR to an extracted bundle path, "+
			"or place the bundle at %s%cruntime%ccv%c "+
			"(see docs/engineering/cv-runtime.md for build instructions)",
		runtime.GOOS, runtime.GOARCH,
		filepath.Dir(exe), filepath.Separator, filepath.Separator, filepath.Separator,
	)
}

// interpreterPath returns the Python interpreter binary inside runtimeDir.
func interpreterPath(runtimeDir string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(runtimeDir, "python", "python.exe")
	}
	return filepath.Join(runtimeDir, "python", "bin", "python3")
}

// entrypointPath returns the reconstruct.py script path inside runtimeDir.
func entrypointPath(runtimeDir string) string {
	return filepath.Join(runtimeDir, "reconstruct.py")
}
