package watcher

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adjust-scans/scanner/internal/logger"
	"github.com/adjust-scans/scanner/internal/processor"
)

func createTestImage(path string) error {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 128, 255})
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
}

func TestNew(t *testing.T) {
	tmpLog := "test_watcher.log"
	defer os.Remove(tmpLog)

	log, err := logger.New(tmpLog)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	tmpProfile := "test_profile.icc"
	os.WriteFile(tmpProfile, []byte("dummy profile data"), 0644)
	defer os.Remove(tmpProfile)

	proc := processor.New(tmpProfile, log)

	testDir := "test_watch"
	outputDir := filepath.Join(testDir, "fixed")
	os.MkdirAll(testDir, 0755)
	os.MkdirAll(outputDir, 0755)
	defer os.RemoveAll(testDir)

	w, err := New(testDir, outputDir, proc, log)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	if w == nil {
		t.Fatal("Watcher is nil")
	}

	w.Stop()
}

func TestWatcherDetectsNewFiles(t *testing.T) {
	tmpLog := "test_watcher.log"
	defer os.Remove(tmpLog)

	log, err := logger.New(tmpLog)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	tmpProfile := "test_profile.icc"
	os.WriteFile(tmpProfile, []byte("dummy profile data"), 0644)
	defer os.Remove(tmpProfile)

	proc := processor.New(tmpProfile, log)

	testDir := "test_watch"
	outputDir := filepath.Join(testDir, "fixed")
	os.MkdirAll(testDir, 0755)
	os.MkdirAll(outputDir, 0755)
	defer os.RemoveAll(testDir)

	w, err := New(testDir, outputDir, proc, log)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer w.Stop()

	// Start watching
	if err := w.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Give watcher time to initialize
	time.Sleep(100 * time.Millisecond)

	// Create a new image file
	testImage := filepath.Join(testDir, "new_test.jpg")
	if err := createTestImage(testImage); err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	// Check if output was created
	outputPath := filepath.Join(outputDir, "new_test.jpg")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Watcher did not process the new file")
	}
}
