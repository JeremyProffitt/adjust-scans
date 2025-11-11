package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Config holds the application configuration
type Config struct {
	ProfilePath string `json:"profile_path"`
	WatchDir    string `json:"watch_dir"`
	OutputDir   string `json:"output_dir"`
	mu          sync.RWMutex
}

var (
	configFile = "scanner_config.json"
)

// Load reads configuration from file or creates default config
func Load() (*Config, error) {
	cfg := &Config{
		OutputDir: "fixed", // Default output directory name
	}

	// Try to load existing config
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, return default config
			return cfg, nil
		}
		return nil, err
	}

	// Parse existing config
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save writes configuration to file
func (c *Config) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}

// GetProfilePath returns the current profile path
func (c *Config) GetProfilePath() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ProfilePath
}

// SetProfilePath sets the profile path and saves config
func (c *Config) SetProfilePath(path string) error {
	c.mu.Lock()
	c.ProfilePath = path
	c.mu.Unlock()
	return c.Save()
}

// GetWatchDir returns the current watch directory
func (c *Config) GetWatchDir() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.WatchDir
}

// SetWatchDir sets the watch directory and saves config
func (c *Config) SetWatchDir(dir string) error {
	c.mu.Lock()
	c.WatchDir = dir
	c.mu.Unlock()
	return c.Save()
}

// GetOutputDir returns the output subdirectory name
func (c *Config) GetOutputDir() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.OutputDir
}

// SetOutputDir sets the output subdirectory name and saves config
func (c *Config) SetOutputDir(dir string) error {
	c.mu.Lock()
	c.OutputDir = dir
	c.mu.Unlock()
	return c.Save()
}

// Validate checks if the configuration is valid for running
func (c *Config) Validate() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var errors []string

	// Check if watch directory is set and exists
	if c.WatchDir == "" {
		errors = append(errors, "Watch directory is not set")
	} else if _, err := os.Stat(c.WatchDir); os.IsNotExist(err) {
		errors = append(errors, "Watch directory does not exist: "+c.WatchDir)
	}

	// Check if profile is set and exists
	if c.ProfilePath == "" {
		errors = append(errors, "ICC profile is not set")
	} else if _, err := os.Stat(c.ProfilePath); os.IsNotExist(err) {
		errors = append(errors, "ICC profile file does not exist: "+c.ProfilePath)
	}

	return errors
}

// GetFullOutputDir returns the full path to the output directory
func (c *Config) GetFullOutputDir() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.WatchDir == "" {
		return ""
	}

	return filepath.Join(c.WatchDir, c.OutputDir)
}
