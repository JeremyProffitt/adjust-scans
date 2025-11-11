# Scanner - Color Profile Image Processor

A lightweight, user-friendly Windows application that automatically applies ICC color profiles to scanned images. Runs quietly in the system tray, providing quick access to all image processing features through an intuitive menu interface.

## Features

### Core Functionality
- **Directory Watching**: Continuously monitors a directory for new TIFF/JPG files and processes them automatically
- **Batch Processing**: Process all images in a directory at once with visual feedback
- **Single File Processing**: Process individual files on-demand through the GUI
- **ICC Color Profile Support**: Apply professional color correction using standard ICC profiles
- **Automatic Output**: Saves color-corrected images to a "fixed" subdirectory (configurable)

### User Interface
- **System Tray Integration**: Runs in the background with a custom scanner icon
- **Interactive Menu**: Right-click access to all features without opening windows
- **File Dialogs**: Easy file and folder selection with native Windows dialogs
- **Auto-Open Results**: Automatically opens processed files/folders when complete
- **Persistent Settings**: Remembers your configuration between sessions

### System Integration
- **Start on Boot**: Optional auto-start when Windows starts
- **Single Instance**: Prevents multiple copies from running simultaneously
- **Background Mode**: Runs silently without console window
- **Comprehensive Logging**: All operations logged for debugging and auditing

## Installation

### Download Pre-built Binary

Download the latest Windows binary from the [Releases](https://github.com/adjust-scans/scanner/releases) page or from the Actions artifacts.

### Build from Source

Requirements:
- Go 1.21 or later
- Git
- Windows (for full GUI functionality)

```bash
git clone https://github.com/adjust-scans/scanner.git
cd scanner
build.bat
```

Or manually:
```bash
go build -v -ldflags="-H=windowsgui" -o scanner.exe .
```

The `-H=windowsgui` flag ensures the application runs in background mode without showing a console window.

## Usage

### Quick Start (Recommended for Most Users)

1. **Launch the Application**
   ```bash
   scanner.exe
   ```

   The application will:
   - Start in the background (no console window)
   - Add a scanner icon to your system tray
   - Return you to the command prompt immediately

2. **Configure Your Settings**

   Right-click the system tray icon and select:
   - **Set Profile** - Choose your ICC color profile file (.icc or .icm)
   - **Set Watch Directory** - Choose a folder to monitor for new scans

3. **Start Processing Images**

   You can now:
   - **Auto-Process**: Drop files in the watch directory - they'll be processed automatically
   - **Manual Process**: Use "Process File" or "Process Directory" from the tray menu
   - **Quick Access**: Use "Open Watch Directory" to quickly navigate to your scan folder

4. **Optional: Enable Auto-Start**

   Right-click the tray icon and check **Start Automatically** to launch Scanner when Windows starts.

### Watch Directory Mode

Monitor a directory for new images and process them automatically:

```bash
scanner.exe -watch "C:\Scans" -profile "C:\Profiles\sRGB.icc"
```

Or, if you've already configured settings via the GUI:

```bash
scanner.exe
```

This will:
- Watch the configured directory for new TIFF/JPG files
- Apply the configured color profile
- Save corrected images to the `fixed` subdirectory
- Show a system tray icon for monitoring

### Batch Process Directory

Process all existing images in a directory and exit:

```bash
scanner.exe -process-dir "C:\Scans" -profile "C:\Profiles\sRGB.icc"
```

### Process Single File

Process one specific file and exit:

```bash
scanner.exe -process-file "C:\Scans\image.tiff" -profile "C:\Profiles\sRGB.icc"
```

### Command Line Options

| Flag | Description | Required |
|------|-------------|----------|
| `-watch` | Directory to watch for new images | No* |
| `-process-dir` | Process all images in directory and exit | No* |
| `-process-file` | Process a specific file and exit | No* |
| `-profile` | Path to ICC color profile file | No** |
| `-output` | Output subdirectory name (default: "fixed") | No |
| `-log` | Log file path (default: "scanner.log") | No |

*If no mode is specified, the application starts in tray mode with GUI configuration.
**Profile is required for `-process-file` and `-process-dir` modes. For watch mode, it can be configured via the Settings menu.

## System Tray Features

Right-click the scanner icon in your system tray to access:

### Configuration
- **Set Profile** - Select an ICC color profile file
  - Opens a file browser filtered to .icc and .icm files
  - Updates take effect immediately (no restart needed)

- **Set Watch Directory** - Choose a directory to monitor
  - Opens a folder browser
  - Requires application restart to begin watching

- **Open Watch Directory** - Opens your configured watch folder in Windows Explorer
  - Quick access to drop/review scans
  - Disabled if no watch directory is configured

### Processing
- **Process File** - Process a single image on-demand
  - Opens a file browser to select an image
  - Applies configured color profile
  - Automatically opens the corrected image when done
  - Saves to "fixed" subfolder in the same directory as the source

- **Process Directory** - Batch process all images in a folder
  - Opens a folder browser to select directory
  - Processes all TIFF and JPEG files
  - Automatically opens the output folder when complete
  - Saves to "fixed" subfolder within the selected directory

### System Options
- **Start Automatically** - Toggle auto-start with Windows
  - Checkmark shows current state
  - Adds/removes Windows Registry startup entry
  - Takes effect on next Windows login

- **Open Log File** - View application log
  - Opens scanner.log in Notepad
  - Useful for troubleshooting

- **Quit** - Exit the application
  - Stops directory watching (if active)
  - Cleanly closes all resources

### Configuration File

The application saves your settings in `scanner_config.json` in the same directory as the executable. This file contains:
- ICC color profile path
- Watch directory path
- Output subdirectory name

You can edit this file manually or use the Settings menu in the system tray.

## Supported Formats

- **Input**: TIFF (.tiff, .tif), JPEG (.jpg, .jpeg)
- **Output**: Same format as input with color profile applied

## Color Profile Notes

The application requires an ICC color profile file. Common profiles include:
- sRGB IEC61966-2.1
- Adobe RGB (1998)
- ProPhoto RGB
- Custom scanner profiles

You can obtain ICC profiles from:
- Your scanner manufacturer
- Color management systems
- Standard profile repositories

### Creating Test Profiles

A Python script is included to generate test ICC profiles:

```bash
python generate_red_profile.py
```

This creates `red_plus_22.icc`, a test profile that increases the red channel by 22 (out of 255). You can use this for testing or as a reference for creating custom profiles.

## Development

### Running Tests

```bash
go test ./...
```

### Project Structure

```
scanner/
├── main.go                    # Application entry point and mode selection
├── build.bat                  # Convenient build script for Windows
├── scanner_config.json        # User configuration (auto-generated)
├── scanner.log                # Application log file (auto-generated)
├── internal/
│   ├── config/               # Configuration file management
│   ├── logger/               # Logging functionality
│   ├── processor/            # Image processing and ICC profile application
│   ├── watcher/              # File system watching
│   ├── tray/                 # System tray icon and menu
│   │   └── scanner_icon.ico  # Embedded tray icon
│   ├── startup/              # Windows auto-start management
│   └── singleton/            # Single instance enforcement
├── .github/
│   └── workflows/
│       └── build.yml         # CI/CD pipeline for Windows builds
├── README.md                 # User documentation (this file)
└── Architecture.md           # Developer documentation
```

For detailed architectural information, see [Architecture.md](Architecture.md).

## License

This project is open source and available under the MIT License.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Troubleshooting

### Application won't start
- **"Another instance is already running"**: Only one copy of Scanner can run at a time
  - Check your system tray for an existing scanner icon
  - Use Task Manager to close any hung scanner.exe processes
  - If problem persists, restart Windows to clear the singleton lock

- **For command-line modes**: Ensure you've specified a valid color profile with `-profile`
- **For GUI mode**: You can configure the profile via the tray menu after starting
- Check that the profile file exists and is readable
- Ensure you're running on Windows (full GUI features require Windows)

### Images not being processed (Auto-Watch Mode)
- Verify the watch directory is configured via "Set Watch Directory"
- Ensure the watch directory exists and is accessible
- Check the log file (Open Log File from tray menu) for error messages
- Verify that a color profile is configured via "Set Profile"
- Ensure files are in TIFF (.tiff, .tif) or JPEG (.jpg, .jpeg) format
- Try manually processing a file via "Process File" to test profile

### Images not being processed (Manual Mode)
- Ensure a color profile is set before using "Process File" or "Process Directory"
- The application will show an error in the log if no profile is configured
- Check file permissions - the application needs read access to source files
- Check folder permissions - the application needs write access to create the "fixed" subfolder

### System tray icon not appearing
- The application always shows a system tray icon when launched
- On Windows, check if hidden tray icons are shown (click the up arrow in system tray)
- If system tray is disabled in Windows, the application may not function correctly
- Try restarting the application

### Auto-start not working
- Verify the "Start Automatically" option is checked in the tray menu
- Check Windows Registry at: `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run`
- The application should have an entry named "Scanner"
- May require administrator privileges on some systems
- Antivirus software may block registry modifications

### Processed images not opening automatically
- Check Windows file associations for image files
- Ensure you have a default program set for TIFF/JPEG files
- Check the log file for errors when attempting to open files/folders
- Manually check the "fixed" subfolder - files may have processed successfully

### Settings not saving
- Check that the application has write permissions in its directory
- The configuration is saved in `scanner_config.json`
- Try running the application from a location where you have full permissions (e.g., Documents folder)
- Avoid running from protected directories (e.g., C:\Program Files\)

### Log file
Check `scanner.log` in the application directory for detailed error messages. This file contains:
- Application startup/shutdown events
- Configuration changes
- Image processing results
- Error messages with details

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/adjust-scans/scanner).
