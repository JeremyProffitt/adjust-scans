//go:build !windows
// +build !windows

package singleton

import (
	"os"
	"path/filepath"
)

var lockFilePath string

// TryLock attempts to acquire the singleton lock
// Returns true if successful, false if another instance is already running
func TryLock() (bool, error) {
	// Use a lock file approach for non-Windows platforms
	tmpDir := os.TempDir()
	lockFilePath = filepath.Join(tmpDir, "scanner.lock")

	// Try to create the lock file exclusively
	file, err := os.OpenFile(lockFilePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		if os.IsExist(err) {
			// Lock file exists - another instance is running
			return false, nil
		}
		return false, err
	}

	// Write PID to lock file
	_, _ = file.WriteString(string(rune(os.Getpid())))
	file.Close()

	return true, nil
}

// Unlock releases the singleton lock
func Unlock() {
	if lockFilePath != "" {
		os.Remove(lockFilePath)
	}
}
