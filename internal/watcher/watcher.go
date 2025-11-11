package watcher

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/adjust-scans/scanner/internal/logger"
	"github.com/adjust-scans/scanner/internal/processor"
	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher   *fsnotify.Watcher
	watchDir  string
	outputDir string
	processor *processor.Processor
	log       *logger.Logger
	done      chan bool
}

func New(watchDir, outputDir string, proc *processor.Processor, log *logger.Logger) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &Watcher{
		watcher:   watcher,
		watchDir:  watchDir,
		outputDir: outputDir,
		processor: proc,
		log:       log,
		done:      make(chan bool),
	}, nil
}

func (w *Watcher) Start() error {
	if err := w.watcher.Add(w.watchDir); err != nil {
		return fmt.Errorf("failed to watch directory: %w", err)
	}

	go w.watch()
	return nil
}

func (w *Watcher) watch() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Only process file creation and write events
			if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
				w.handleFileEvent(event.Name)
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			w.log.Errorf("Watcher error: %v", err)

		case <-w.done:
			return
		}
	}
}

func (w *Watcher) handleFileEvent(filePath string) {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Check if it's a supported image format
	if ext == ".tiff" || ext == ".tif" || ext == ".jpg" || ext == ".jpeg" {
		w.log.Infof("New file detected: %s", filePath)

		// Process the image
		if err := w.processor.ProcessImage(filePath, w.outputDir); err != nil {
			w.log.Errorf("Failed to process %s: %v", filePath, err)
		}
	}
}

func (w *Watcher) Stop() {
	close(w.done)
	w.watcher.Close()
	w.log.Info("Watcher stopped")
}
