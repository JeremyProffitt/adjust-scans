package tray

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/adjust-scans/scanner/internal/logger"
	"github.com/adjust-scans/scanner/internal/processor"
	"github.com/getlantern/systray"
)

type Tray struct {
	log         *logger.Logger
	processor   *processor.Processor
	logFilePath string
}

func New(log *logger.Logger, proc *processor.Processor, logFilePath string) (*Tray, error) {
	return &Tray{
		log:         log,
		processor:   proc,
		logFilePath: logFilePath,
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
	mOpenLog := systray.AddMenuItem("Open Log File", "Open the log file")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	// Handle menu clicks
	go func() {
		for {
			select {
			case <-mRecent.ClickedCh:
				t.showRecentImages()

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
