package startup

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const (
	registryPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	appName      = "Scanner"
)

// IsEnabled checks if the application is set to start automatically
func IsEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, registryPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	_, _, err = k.GetStringValue(appName)
	return err == nil
}

// Enable adds the application to Windows startup
func Enable() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	// Convert to absolute path
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		return err
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, registryPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	return k.SetStringValue(appName, exePath)
}

// Disable removes the application from Windows startup
func Disable() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, registryPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	return k.DeleteValue(appName)
}
