package tray

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/adjust-scans/scanner/internal/config"
	"github.com/adjust-scans/scanner/internal/logger"
	"github.com/adjust-scans/scanner/internal/processor"
	"github.com/getlantern/systray"
	"github.com/sqweek/dialog"
)

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
	systray.SetTitle("Scanner")
	systray.SetTooltip("Image Color Profile Scanner")

	// Create menu items
	mRecent := systray.AddMenuItem("Recent Images", "View recently processed images")
	systray.AddSeparator()
	mSettings := systray.AddMenuItem("Settings", "Configure scanner settings")
	mOpenLog := systray.AddMenuItem("Open Log File", "Open the log file")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	// Handle menu clicks
	go func() {
		for {
			select {
			case <-mRecent.ClickedCh:
				t.showRecentImages()

			case <-mSettings.ClickedCh:
				t.showSettings()

			case <-mOpenLog.ClickedCh:
				t.openLogFile()

			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()

	// Update recent images periodically
	go t.updateRecentImagesMenu(mRecent)
}

func (t *Tray) onExit() {
	t.log.Info("Application exiting")
}

func (t *Tray) updateRecentImagesMenu(menu *systray.MenuItem) {
	// This would be called periodically to update the submenu
	// For simplicity, we'll just log the recent images when clicked
}

func (t *Tray) showRecentImages() {
	images := t.processor.GetRecentImages()

	if len(images) == 0 {
		t.log.Info("No recent images to display")
		return
	}

	message := "Recent Images (Last 10):\n\n"
	for i, img := range images {
		status := "SUCCESS"
		if !img.Success {
			status = fmt.Sprintf("FAILED: %s", img.Error)
		}
		message += fmt.Sprintf("%d. %s - %s\n   Time: %s\n",
			i+1, img.FileName, status, img.ProcessedTime.Format("2006-01-02 15:04:05"))
	}

	t.log.Info("Recent images requested")
	fmt.Println("\n" + message)
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
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
}

func (t *Tray) showSettings() {
	t.log.Info("Settings menu opened")

	// Show current settings
	currentProfile := t.config.GetProfilePath()
	currentWatchDir := t.config.GetWatchDir()

	if currentProfile == "" {
		currentProfile = "(not set)"
	}
	if currentWatchDir == "" {
		currentWatchDir = "(not set)"
	}

	fmt.Println("\n=== Current Settings ===")
	fmt.Printf("ICC Profile: %s\n", currentProfile)
	fmt.Printf("Watch Directory: %s\n", currentWatchDir)
	fmt.Printf("Output Directory: %s\n", t.config.GetOutputDir())
	fmt.Println("\nOptions:")
	fmt.Println("1. Set ICC Profile")
	fmt.Println("2. Set Watch Directory")
	fmt.Println()

	// Note: In a real GUI, you'd show a proper dialog
	// For now, we'll provide a way to set via file dialogs
	t.promptSettings()
}

func (t *Tray) promptSettings() {
	profileUpdated := false
	watchDirUpdated := false

	// Prompt for ICC profile
	if profile, err := dialog.File().
		Title("Select ICC Profile (or Cancel to skip)").
		Filter("ICC Profiles", "icc", "icm").
		Load(); err == nil && profile != "" {

		if err := t.config.SetProfilePath(profile); err != nil {
			t.log.Errorf("Failed to save profile path: %v", err)
			fmt.Printf("Error saving profile: %v\n", err)
		} else {
			t.log.Infof("Profile set to: %s", profile)
			fmt.Printf("Profile updated: %s\n", profile)

			// Update processor with new profile immediately
			if err := t.processor.UpdateProfile(profile); err != nil {
				t.log.Errorf("Failed to update processor profile: %v", err)
			} else {
				profileUpdated = true
			}
		}
	}

	// Prompt for watch directory
	if dir, err := dialog.Directory().
		Title("Select Watch Directory (or Cancel to skip)").
		Browse(); err == nil && dir != "" {

		if err := t.config.SetWatchDir(dir); err != nil {
			t.log.Errorf("Failed to save watch directory: %v", err)
			fmt.Printf("Error saving watch directory: %v\n", err)
		} else {
			t.log.Infof("Watch directory set to: %s", dir)
			fmt.Printf("Watch directory updated: %s\n", dir)
			watchDirUpdated = true
		}
	}

	// Notify about config changes
	if (profileUpdated || watchDirUpdated) && t.onConfigChange != nil {
		t.onConfigChange()
	}

	if watchDirUpdated {
		fmt.Println("\nSettings updated. Restart the application to watch the new directory.")
	} else if profileUpdated {
		fmt.Println("\nSettings updated. New profile is now active.")
	}
}
