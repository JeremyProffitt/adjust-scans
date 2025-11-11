package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

type Logger struct {
	file   *os.File
	logger *log.Logger
	mu     sync.Mutex
}

type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
}

var (
	recentLogs   []LogEntry
	recentLogsMu sync.Mutex
	maxRecentLog = 100
)

func New(filePath string) (*Logger, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	multiWriter := io.MultiWriter(file, os.Stdout)
	logger := log.New(multiWriter, "", log.LstdFlags)

	return &Logger{
		file:   file,
		logger: logger,
	}, nil
}

func (l *Logger) log(level, format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	message := fmt.Sprintf(format, v...)
	l.logger.Printf("[%s] %s", level, message)

	// Store in recent logs
	recentLogsMu.Lock()
	recentLogs = append(recentLogs, LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
	})
	if len(recentLogs) > maxRecentLog {
		recentLogs = recentLogs[1:]
	}
	recentLogsMu.Unlock()
}

func (l *Logger) Info(msg string) {
	l.log("INFO", "%s", msg)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.log("INFO", format, v...)
}

func (l *Logger) Error(msg string) {
	l.log("ERROR", "%s", msg)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.log("ERROR", format, v...)
}

func (l *Logger) Warning(msg string) {
	l.log("WARNING", "%s", msg)
}

func (l *Logger) Warningf(format string, v ...interface{}) {
	l.log("WARNING", format, v...)
}

func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func GetRecentLogs() []LogEntry {
	recentLogsMu.Lock()
	defer recentLogsMu.Unlock()

	// Return a copy
	logs := make([]LogEntry, len(recentLogs))
	copy(logs, recentLogs)
	return logs
}
