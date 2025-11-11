package logger

import (
	"os"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tmpFile := "test_log.log"
	defer os.Remove(tmpFile)

	log, err := New(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	if log == nil {
		t.Fatal("Logger is nil")
	}
}

func TestLogging(t *testing.T) {
	tmpFile := "test_log.log"
	defer os.Remove(tmpFile)

	log, err := New(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	// Test different log levels
	log.Info("Test info message")
	log.Infof("Test info with format: %d", 123)
	log.Error("Test error message")
	log.Errorf("Test error with format: %s", "error")
	log.Warning("Test warning message")
	log.Warningf("Test warning with format: %v", true)

	// Verify log file was created and has content
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "Test info message") {
		t.Error("Log file doesn't contain info message")
	}
	if !strings.Contains(content, "Test error message") {
		t.Error("Log file doesn't contain error message")
	}
	if !strings.Contains(content, "Test warning message") {
		t.Error("Log file doesn't contain warning message")
	}
}

func TestRecentLogs(t *testing.T) {
	tmpFile := "test_log.log"
	defer os.Remove(tmpFile)

	log, err := New(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	// Log some messages with unique identifiers
	log.Info("UniqueMessage1")
	log.Info("UniqueMessage2")
	log.Error("UniqueMessage3")

	// Get recent logs
	logs := GetRecentLogs()

	// Verify we have at least 3 logs
	if len(logs) < 3 {
		t.Errorf("Expected at least 3 recent logs, got %d", len(logs))
	}

	// Verify our messages exist in the recent logs
	found := make(map[string]bool)
	for _, entry := range logs {
		if strings.Contains(entry.Message, "UniqueMessage1") {
			found["UniqueMessage1"] = true
		}
		if strings.Contains(entry.Message, "UniqueMessage2") {
			found["UniqueMessage2"] = true
		}
		if strings.Contains(entry.Message, "UniqueMessage3") {
			found["UniqueMessage3"] = true
		}
	}

	if !found["UniqueMessage1"] || !found["UniqueMessage2"] || !found["UniqueMessage3"] {
		t.Error("Not all logged messages were found in recent logs")
	}
}
