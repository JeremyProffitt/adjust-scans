//go:build !windows
// +build !windows

package startup

import "errors"

// IsEnabled checks if the application is set to start automatically
func IsEnabled() bool {
	return false
}

// Enable adds the application to startup (not implemented for non-Windows)
func Enable() error {
	return errors.New("auto-start is only supported on Windows")
}

// Disable removes the application from startup (not implemented for non-Windows)
func Disable() error {
	return errors.New("auto-start is only supported on Windows")
}
