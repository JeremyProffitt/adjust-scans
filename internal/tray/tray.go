package tray

import (
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/adjust-scans/scanner/internal/config"
	"github.com/adjust-scans/scanner/internal/logger"
	"github.com/adjust-scans/scanner/internal/processor"
	"github.com/adjust-scans/scanner/internal/startup"
	"github.com/getlantern/systray"
	"github.com/sqweek/dialog"
)

//go:embed scanner_icon.ico
var iconData []byte

type Tray struct {
	log         *logger.Logger
	processor   *processor.Processor
	config      *config.Config
	logFilePath string
	onConfigChange func()
}

func New(log *logger.Logger, proc *processor.Processor, cfg *config.Config, logFilePath string, onConfigChange func()) (*Tray, error) {
	return &Tray{
		log:         log,
		processor:   proc,
		config:      cfg,
		logFilePath: logFilePath,
		onConfigChange: onConfigChange,
	}, nil
}

func (t *Tray) Run() {
	systray.Run(t.onReady, t.onExit)
}

func (t *Tray) onReady() {
	// Set the icon
	systray.SetIcon(iconData)
	systray.SetTitle("Scanner")
	systray.SetTooltip("Image Color Profile Scanner")

	// Create menu items
	mSetProfile := systray.AddMenuItem("Set Profile", "Select ICC color profile")
	mSetWatchDir := systray.AddMenuItem("Set Watch Directory", "Select directory to watch for new images")
	mOpenWatchDir := systray.AddMenuItem("Open Watch Directory", "Open the watch directory in file explorer")
	systray.AddSeparator()
	mProcessFile := systray.AddMenuItem("Process File", "Process a single image file")
	mProcessDir := systray.AddMenuItem("Process Directory", "Process all images in a directory")
	systray.AddSeparator()
	mStartAuto := systray.AddMenuItemCheckbox("Start Automatically", "Start scanner when Windows starts", startup.IsEnabled())
	systray.AddSeparator()
	mOpenLog := systray.AddMenuItem("Open Log File", "Open the log file")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	// Handle menu clicks
	go func() {
		for {
			select {
			case <-mSetProfile.ClickedCh:
				t.setProfile()

			case <-mSetWatchDir.ClickedCh:
				t.setWatchDirectory()

			case <-mOpenWatchDir.ClickedCh:
				t.openWatchDirectory()

			case <-mProcessFile.ClickedCh:
				t.processFile()

			case <-mProcessDir.ClickedCh:
				t.processDirectory()

			case <-mStartAuto.ClickedCh:
				t.toggleStartAutomatically(mStartAuto)

			case <-mOpenLog.ClickedCh:
				t.openLogFile()

			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func (t *Tray) onExit() {
	t.log.Info("Application exiting")
}

func (t *Tray) toggleStartAutomatically(menuItem *systray.MenuItem) {
	if startup.IsEnabled() {
		// Disable startup
		if err := startup.Disable(); err != nil {
			t.log.Errorf("Failed to disable auto-start: %v", err)
			return
		}
		menuItem.Uncheck()
		t.log.Info("Auto-start disabled")
	} else {
		// Enable startup
		if err := startup.Enable(); err != nil {
			t.log.Errorf("Failed to enable auto-start: %v", err)
			return
		}
		menuItem.Check()
		t.log.Info("Auto-start enabled")
	}
}

func (t *Tray) setProfile() {
	t.log.Info("Set profile menu clicked")

	profile, err := dialog.File().
		Title("Select ICC Profile").
		Filter("ICC Profiles", "icc", "icm").
		Load()

	if err != nil {
		if err.Error() != "Cancelled" {
			t.log.Errorf("Error selecting profile: %v", err)
		}
		return
	}

	if profile == "" {
		return
	}

	if err := t.config.SetProfilePath(profile); err != nil {
		t.log.Errorf("Failed to save profile path: %v", err)
		return
	}

	t.log.Infof("Profile set to: %s", profile)

	// Update processor with new profile immediately
	if err := t.processor.UpdateProfile(profile); err != nil {
		t.log.Errorf("Failed to update processor profile: %v", err)
	}

	if t.onConfigChange != nil {
		t.onConfigChange()
	}
}

func (t *Tray) setWatchDirectory() {
	t.log.Info("Set watch directory menu clicked")

	dir, err := dialog.Directory().
		Title("Select Watch Directory").
		Browse()

	if err != nil {
		if err.Error() != "Cancelled" {
			t.log.Errorf("Error selecting directory: %v", err)
		}
		return
	}

	if dir == "" {
		return
	}

	if err := t.config.SetWatchDir(dir); err != nil {
		t.log.Errorf("Failed to save watch directory: %v", err)
		return
	}

	t.log.Infof("Watch directory set to: %s", dir)

	if t.onConfigChange != nil {
		t.onConfigChange()
	}
}

func (t *Tray) openWatchDirectory() {
	watchDir := t.config.GetWatchDir()

	if watchDir == "" {
		t.log.Info("No watch directory configured")
		return
	}

	t.log.Infof("Opening watch directory: %s", watchDir)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", watchDir)
	case "darwin":
		cmd = exec.Command("open", watchDir)
	default:
		cmd = exec.Command("xdg-open", watchDir)
	}

	if err := cmd.Start(); err != nil {
		t.log.Errorf("Failed to open directory: %v", err)
	}
}

func (t *Tray) processFile() {
	t.log.Info("Process file menu clicked")

	// Check if profile is configured
	if t.config.GetProfilePath() == "" {
		t.log.Error("No profile configured - please set profile first")
		return
	}

	file, err := dialog.File().
		Title("Select Image File to Process").
		Filter("Image Files", "tiff", "tif", "jpg", "jpeg").
		Load()

	if err != nil {
		if err.Error() != "Cancelled" {
			t.log.Errorf("Error selecting file: %v", err)
		}
		return
	}

	if file == "" {
		return
	}

	t.log.Infof("Processing file: %s", file)

	dir := filepath.Dir(file)
	outputDir := filepath.Join(dir, t.config.GetOutputDir())

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.log.Errorf("Failed to create output directory: %v", err)
		return
	}

	if err := t.processor.ProcessImage(file, outputDir); err != nil {
		t.log.Errorf("Failed to process file: %v", err)
	} else {
		t.log.Infof("File processed successfully: %s", file)

		// Open the output file
		outputFile := filepath.Join(outputDir, filepath.Base(file))
		t.openFile(outputFile)
	}
}

func (t *Tray) openFile(filePath string) {
	t.log.Infof("Opening file: %s", filePath)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", filePath)
	case "darwin":
		cmd = exec.Command("open", filePath)
	default:
		cmd = exec.Command("xdg-open", filePath)
	}

	if err := cmd.Start(); err != nil {
		t.log.Errorf("Failed to open file: %v", err)
	}
}

func (t *Tray) processDirectory() {
	t.log.Info("Process directory menu clicked")

	// Check if profile is configured
	if t.config.GetProfilePath() == "" {
		t.log.Error("No profile configured - please set profile first")
		return
	}

	dir, err := dialog.Directory().
		Title("Select Directory to Process").
		Browse()

	if err != nil {
		if err.Error() != "Cancelled" {
			t.log.Errorf("Error selecting directory: %v", err)
		}
		return
	}

	if dir == "" {
		return
	}

	t.log.Infof("Processing directory: %s", dir)

	outputDir := filepath.Join(dir, t.config.GetOutputDir())

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.log.Errorf("Failed to create output directory: %v", err)
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.log.Errorf("Failed to read directory: %v", err)
		return
	}

	processed := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		ext := filepath.Ext(filePath)

		if ext == ".tiff" || ext == ".tif" || ext == ".jpg" || ext == ".jpeg" {
			t.log.Infof("Processing: %s", filePath)
			if err := t.processor.ProcessImage(filePath, outputDir); err != nil {
				t.log.Errorf("Failed to process %s: %v", filePath, err)
				continue
			}
			processed++
		}
	}

	t.log.Infof("Processed %d images from directory: %s", processed, dir)

	// Open the output folder
	if processed > 0 {
		t.openFolder(outputDir)
	}
}

func (t *Tray) openFolder(folderPath string) {
	t.log.Infof("Opening folder: %s", folderPath)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", folderPath)
	case "darwin":
		cmd = exec.Command("open", folderPath)
	default:
		cmd = exec.Command("xdg-open", folderPath)
	}

	if err := cmd.Start(); err != nil {
		t.log.Errorf("Failed to open folder: %v", err)
	}
}

func (t *Tray) openLogFile() {
	t.log.Info("Opening log file")

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("notepad.exe", t.logFilePath)
	case "darwin":
		cmd = exec.Command("open", t.logFilePath)
	default:
		cmd = exec.Command("xdg-open", t.logFilePath)
	}

	if err := cmd.Start(); err != nil {
		t.log.Errorf("Failed to open log file: %v", err)
	}
}
