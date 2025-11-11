package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adjust-scans/scanner/internal/config"
	"github.com/adjust-scans/scanner/internal/logger"
	"github.com/adjust-scans/scanner/internal/processor"
	"github.com/adjust-scans/scanner/internal/singleton"
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

	// Check for singleton instance
	locked, err := singleton.TryLock()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to check for existing instance: %v\n", err)
		os.Exit(1)
	}
	if !locked {
		fmt.Fprintln(os.Stderr, "Another instance of Scanner is already running")
		os.Exit(1)
	}
	defer singleton.Unlock()

	// Initialize logger
	log, err := logger.New(*logFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Info("Scanner application started")

	// Load or create configuration
	cfg, err := config.Load()
	if err != nil {
		log.Errorf("Failed to load configuration: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Override config with command-line flags if provided
	if *colorProfile != "" {
		cfg.SetProfilePath(*colorProfile)
	}
	if *watchDir != "" {
		cfg.SetWatchDir(*watchDir)
	}
	if *outputDir != "" {
		cfg.SetOutputDir(*outputDir)
	}

	// Create processor (will work even without profile, but won't apply corrections)
	profilePath := cfg.GetProfilePath()
	if *colorProfile != "" {
		profilePath = *colorProfile
	}
	proc := processor.New(profilePath, log)

	// Handle different modes
	switch {
	case *processFile != "":
		// Process single file and exit
		if profilePath == "" {
			log.Error("Color profile is required for processing. Use -profile flag or configure in settings")
			fmt.Fprintln(os.Stderr, "Error: Color profile is required. Use -profile flag")
			os.Exit(1)
		}
		if err := processSingleFile(proc, *processFile, log); err != nil {
			log.Errorf("Failed to process file: %v", err)
			os.Exit(1)
		}
		log.Info("File processed successfully")

	case *processDir != "":
		// Process all files in directory and exit
		if profilePath == "" {
			log.Error("Color profile is required for processing. Use -profile flag or configure in settings")
			fmt.Fprintln(os.Stderr, "Error: Color profile is required. Use -profile flag")
			os.Exit(1)
		}
		outputDirName := *outputDir
		if outputDirName == "" {
			outputDirName = cfg.GetOutputDir()
		}
		if err := processDirectory(proc, *processDir, outputDirName, log); err != nil {
			log.Errorf("Failed to process directory: %v", err)
			os.Exit(1)
		}
		log.Info("Directory processed successfully")

	case *watchDir != "" || cfg.GetWatchDir() != "":
		// Watch directory mode
		watchDirPath := *watchDir
		if watchDirPath == "" {
			watchDirPath = cfg.GetWatchDir()
		}
		outputDirName := *outputDir
		if outputDirName == "" {
			outputDirName = cfg.GetOutputDir()
		}
		if err := watchDirectory(proc, cfg, watchDirPath, outputDirName, log); err != nil {
			log.Errorf("Failed to start watcher: %v", err)
			os.Exit(1)
		}

	default:
		// No mode specified - start in tray mode
		log.Info("Starting in tray mode")
		if err := trayMode(proc, cfg, log); err != nil {
			log.Errorf("Failed to start tray mode: %v", err)
			os.Exit(1)
		}
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

func watchDirectory(proc *processor.Processor, cfg *config.Config, dirPath, outputSubdir string, log *logger.Logger) error {
	outputDir := filepath.Join(dirPath, outputSubdir)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create watcher
	w, err := watcher.New(dirPath, outputDir, proc, log)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	// Start watcher
	if err := w.Start(); err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	log.Infof("Watching directory: %s", dirPath)
	log.Infof("Output directory: %s", outputDir)

	// Create system tray with config change handler
	onConfigChange := func() {
		log.Info("Configuration changed - restart application to apply changes")
	}

	t, err := tray.New(log, proc, cfg, *logFile, onConfigChange)
	if err != nil {
		return fmt.Errorf("failed to create system tray: %w", err)
	}

	// Run system tray (blocks until quit)
	t.Run()

	// Cleanup
	w.Stop()
	return nil
}

func trayMode(proc *processor.Processor, cfg *config.Config, log *logger.Logger) error {
	// Create system tray with config change handler
	onConfigChange := func() {
		log.Info("Configuration changed - settings saved")
	}

	t, err := tray.New(log, proc, cfg, *logFile, onConfigChange)
	if err != nil {
		return fmt.Errorf("failed to create system tray: %w", err)
	}

	// Run system tray (blocks until quit)
	t.Run()
	return nil
}
