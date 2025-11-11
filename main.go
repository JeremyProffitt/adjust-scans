package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adjust-scans/scanner/internal/logger"
	"github.com/adjust-scans/scanner/internal/processor"
	"github.com/adjust-scans/scanner/internal/tray"
	"github.com/adjust-scans/scanner/internal/watcher"
)

var (
	watchDir      = flag.String("watch", "", "Directory to watch for new images")
	processDir    = flag.String("process-dir", "", "Process all images in directory and exit")
	processFile   = flag.String("process-file", "", "Process a specific file and exit")
	colorProfile  = flag.String("profile", "", "Path to ICC color profile file")
	outputDir     = flag.String("output", "fixed", "Output subdirectory name (default: fixed)")
	logFile       = flag.String("log", "scanner.log", "Log file path")
)

func main() {
	flag.Parse()

	// Initialize logger
	log, err := logger.New(*logFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Info("Scanner application started")

	// Validate color profile
	if *colorProfile == "" {
		log.Error("Color profile is required. Use -profile flag")
		fmt.Fprintln(os.Stderr, "Error: Color profile is required. Use -profile flag")
		os.Exit(1)
	}

	if _, err := os.Stat(*colorProfile); os.IsNotExist(err) {
		log.Errorf("Color profile file not found: %s", *colorProfile)
		fmt.Fprintf(os.Stderr, "Error: Color profile file not found: %s\n", *colorProfile)
		os.Exit(1)
	}

	// Create processor
	proc := processor.New(*colorProfile, log)

	// Handle different modes
	switch {
	case *processFile != "":
		// Process single file and exit
		if err := processSingleFile(proc, *processFile, log); err != nil {
			log.Errorf("Failed to process file: %v", err)
			os.Exit(1)
		}
		log.Info("File processed successfully")

	case *processDir != "":
		// Process all files in directory and exit
		if err := processDirectory(proc, *processDir, *outputDir, log); err != nil {
			log.Errorf("Failed to process directory: %v", err)
			os.Exit(1)
		}
		log.Info("Directory processed successfully")

	case *watchDir != "":
		// Watch directory mode
		if err := watchDirectory(proc, *watchDir, *outputDir, log); err != nil {
			log.Errorf("Failed to start watcher: %v", err)
			os.Exit(1)
		}

	default:
		flag.Usage()
		fmt.Fprintln(os.Stderr, "\nError: You must specify one of: -watch, -process-dir, or -process-file")
		os.Exit(1)
	}
}

func processSingleFile(proc *processor.Processor, filePath string, log *logger.Logger) error {
	dir := filepath.Dir(filePath)
	outputDir := filepath.Join(dir, *outputDir)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return proc.ProcessImage(filePath, outputDir)
}

func processDirectory(proc *processor.Processor, dirPath, outputSubdir string, log *logger.Logger) error {
	outputDir := filepath.Join(dirPath, outputSubdir)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	processed := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		ext := filepath.Ext(filePath)

		if ext == ".tiff" || ext == ".tif" || ext == ".jpg" || ext == ".jpeg" {
			log.Infof("Processing: %s", filePath)
			if err := proc.ProcessImage(filePath, outputDir); err != nil {
				log.Errorf("Failed to process %s: %v", filePath, err)
				continue
			}
			processed++
		}
	}

	log.Infof("Processed %d images", processed)
	return nil
}

func watchDirectory(proc *processor.Processor, dirPath, outputSubdir string, log *logger.Logger) error {
	outputDir := filepath.Join(dirPath, outputSubdir)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create watcher
	w, err := watcher.New(dirPath, outputDir, proc, log)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	// Create system tray
	t, err := tray.New(log, proc, *logFile)
	if err != nil {
		return fmt.Errorf("failed to create system tray: %w", err)
	}

	// Start watcher
	if err := w.Start(); err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	log.Infof("Watching directory: %s", dirPath)
	log.Infof("Output directory: %s", outputDir)

	// Run system tray (blocks until quit)
	t.Run()

	// Cleanup
	w.Stop()
	return nil
}
