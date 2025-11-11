package processor

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/adjust-scans/scanner/internal/logger"
	"github.com/disintegration/imaging"
	"golang.org/x/image/tiff"
)

type ProcessedImage struct {
	FileName      string
	ProcessedTime time.Time
	InputPath     string
	OutputPath    string
	Success       bool
	Error         string
}

type Processor struct {
	profilePath    string
	profileData    []byte
	log            *logger.Logger
	recentImages   []ProcessedImage
	recentImagesMu sync.RWMutex
	maxRecent      int
}

func New(profilePath string, log *logger.Logger) *Processor {
	// Read color profile
	profileData, err := os.ReadFile(profilePath)
	if err != nil {
		log.Errorf("Failed to read color profile: %v", err)
		profileData = nil
	}

	return &Processor{
		profilePath: profilePath,
		profileData: profileData,
		log:         log,
		maxRecent:   10,
	}
}

func (p *Processor) ProcessImage(inputPath, outputDir string) error {
	startTime := time.Now()
	fileName := filepath.Base(inputPath)
	outputPath := filepath.Join(outputDir, fileName)

	p.log.Infof("Processing image: %s", inputPath)

	// Record processing attempt
	processedImg := ProcessedImage{
		FileName:      fileName,
		ProcessedTime: startTime,
		InputPath:     inputPath,
		OutputPath:    outputPath,
		Success:       false,
	}

	defer func() {
		p.addRecentImage(processedImg)
	}()

	// Open input image
	img, format, err := p.loadImage(inputPath)
	if err != nil {
		processedImg.Error = fmt.Sprintf("Failed to load image: %v", err)
		p.log.Errorf("Failed to load image %s: %v", inputPath, err)
		return err
	}

	// Apply color profile correction
	correctedImg, err := p.applyColorProfile(img)
	if err != nil {
		processedImg.Error = fmt.Sprintf("Failed to apply color profile: %v", err)
		p.log.Errorf("Failed to apply color profile to %s: %v", inputPath, err)
		return err
	}

	// Save output image
	if err := p.saveImage(correctedImg, outputPath, format); err != nil {
		processedImg.Error = fmt.Sprintf("Failed to save image: %v", err)
		p.log.Errorf("Failed to save image %s: %v", outputPath, err)
		return err
	}

	processedImg.Success = true
	p.log.Infof("Successfully processed %s -> %s (%.2fs)", inputPath, outputPath, time.Since(startTime).Seconds())

	return nil
}

func (p *Processor) loadImage(path string) (image.Image, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	ext := filepath.Ext(path)
	var img image.Image

	switch ext {
	case ".tiff", ".tif":
		img, err = tiff.Decode(file)
		if err != nil {
			return nil, "", fmt.Errorf("failed to decode TIFF: %w", err)
		}
		return img, "tiff", nil

	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
		if err != nil {
			return nil, "", fmt.Errorf("failed to decode JPEG: %w", err)
		}
		return img, "jpeg", nil

	default:
		return nil, "", fmt.Errorf("unsupported format: %s", ext)
	}
}

func (p *Processor) applyColorProfile(img image.Image) (image.Image, error) {
	// For now, we'll apply basic color correction
	// In a production environment, you would use a proper ICC profile library
	// such as github.com/mandykoh/prism for full ICC profile support

	// Apply contrast and brightness adjustments as a placeholder for color profile correction
	// This simulates the color correction that would be done by a proper ICC profile
	adjusted := imaging.AdjustSaturation(img, 10)
	adjusted = imaging.AdjustContrast(adjusted, 5)

	return adjusted, nil
}

func (p *Processor) saveImage(img image.Image, path, format string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	switch format {
	case "tiff":
		return tiff.Encode(file, img, &tiff.Options{
			Compression: tiff.Deflate,
		})

	case "jpeg":
		return jpeg.Encode(file, img, &jpeg.Options{
			Quality: 95,
		})

	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func (p *Processor) addRecentImage(img ProcessedImage) {
	p.recentImagesMu.Lock()
	defer p.recentImagesMu.Unlock()

	p.recentImages = append([]ProcessedImage{img}, p.recentImages...)
	if len(p.recentImages) > p.maxRecent {
		p.recentImages = p.recentImages[:p.maxRecent]
	}
}

func (p *Processor) GetRecentImages() []ProcessedImage {
	p.recentImagesMu.RLock()
	defer p.recentImagesMu.RUnlock()

	images := make([]ProcessedImage, len(p.recentImages))
	copy(images, p.recentImages)
	return images
}
