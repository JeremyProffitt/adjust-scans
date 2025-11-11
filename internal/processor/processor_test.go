package processor

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/adjust-scans/scanner/internal/logger"
)

func createTestImage(path string) error {
	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Fill with a gradient
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
	tmpLog := "test_processor.log"
	defer os.Remove(tmpLog)

	log, err := logger.New(tmpLog)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	// Create a dummy profile file
	tmpProfile := "test_profile.icc"
	os.WriteFile(tmpProfile, []byte("dummy profile data"), 0644)
	defer os.Remove(tmpProfile)

	proc := New(tmpProfile, log)
	if proc == nil {
		t.Fatal("Processor is nil")
	}
}

func TestProcessImage(t *testing.T) {
	tmpLog := "test_processor.log"
	defer os.Remove(tmpLog)

	log, err := logger.New(tmpLog)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	// Create a dummy profile file
	tmpProfile := "test_profile.icc"
	os.WriteFile(tmpProfile, []byte("dummy profile data"), 0644)
	defer os.Remove(tmpProfile)

	proc := New(tmpProfile, log)

	// Create test directories
	testDir := "test_images"
	outputDir := filepath.Join(testDir, "fixed")
	os.MkdirAll(outputDir, 0755)
	defer os.RemoveAll(testDir)

	// Create test image
	testImage := filepath.Join(testDir, "test.jpg")
	if err := createTestImage(testImage); err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	// Process the image
	if err := proc.ProcessImage(testImage, outputDir); err != nil {
		t.Fatalf("Failed to process image: %v", err)
	}

	// Verify output exists
	outputPath := filepath.Join(outputDir, "test.jpg")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output image was not created")
	}

	// Verify recent images tracking
	recent := proc.GetRecentImages()
	if len(recent) != 1 {
		t.Errorf("Expected 1 recent image, got %d", len(recent))
	}

	if !recent[0].Success {
		t.Errorf("Image processing should have succeeded")
	}
}

func TestLoadImage(t *testing.T) {
	tmpLog := "test_processor.log"
	defer os.Remove(tmpLog)

	log, err := logger.New(tmpLog)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	tmpProfile := "test_profile.icc"
	os.WriteFile(tmpProfile, []byte("dummy profile data"), 0644)
	defer os.Remove(tmpProfile)

	proc := New(tmpProfile, log)

	// Create test image
	testImage := "test_load.jpg"
	if err := createTestImage(testImage); err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}
	defer os.Remove(testImage)

	// Test loading
	img, format, err := proc.loadImage(testImage)
	if err != nil {
		t.Fatalf("Failed to load image: %v", err)
	}

	if img == nil {
		t.Fatal("Loaded image is nil")
	}

	if format != "jpeg" {
		t.Errorf("Expected format 'jpeg', got '%s'", format)
	}
}

func TestGetRecentImages(t *testing.T) {
	tmpLog := "test_processor.log"
	defer os.Remove(tmpLog)

	log, err := logger.New(tmpLog)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	tmpProfile := "test_profile.icc"
	os.WriteFile(tmpProfile, []byte("dummy profile data"), 0644)
	defer os.Remove(tmpProfile)

	proc := New(tmpProfile, log)

	// Initially should be empty
	recent := proc.GetRecentImages()
	if len(recent) != 0 {
		t.Errorf("Expected 0 recent images, got %d", len(recent))
	}

	// Process multiple images
	testDir := "test_images"
	outputDir := filepath.Join(testDir, "fixed")
	os.MkdirAll(outputDir, 0755)
	defer os.RemoveAll(testDir)

	for i := 0; i < 5; i++ {
		testImage := filepath.Join(testDir, "test"+string(rune('0'+i))+".jpg")
		createTestImage(testImage)
		proc.ProcessImage(testImage, outputDir)
	}

	recent = proc.GetRecentImages()
	if len(recent) != 5 {
		t.Errorf("Expected 5 recent images, got %d", len(recent))
	}
}
