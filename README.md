# Scanner - Color Profile Image Processor

A Go application that automatically applies ICC color profiles to scanned images (TIFF and JPEG formats).

## Features

- **Directory Watching**: Continuously monitors a directory for new TIFF/JPG files
- **Batch Processing**: Process all images in a directory at once
- **Single File Processing**: Process individual files
- **System Tray Icon**: View last 10 processed images and open log file
- **Automatic Output**: Saves color-corrected images to a "fixed" subdirectory
- **Comprehensive Logging**: All operations are logged for debugging and auditing

## Installation

### Download Pre-built Binary

Download the latest Windows binary from the [Releases](https://github.com/adjust-scans/scanner/releases) page or from the Actions artifacts.

### Build from Source

Requirements:
- Go 1.21 or later
- Git

```bash
git clone https://github.com/adjust-scans/scanner.git
cd scanner
go build -o scanner.exe .
```

## Usage

### Quick Start (GUI Configuration)

Simply run the scanner without any arguments to start in tray mode:

```bash
scanner.exe
```

This will:
- Show a system tray icon
- Allow you to configure settings via the right-click menu
- Save your configuration for future use

Right-click the tray icon and select **Settings** to:
- Choose your ICC color profile
- Select the directory to watch for new images

Once configured, restart the application to start watching for new images automatically.

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

When running the application, the system tray icon provides:

- **Recent Images**: View the last 10 processed images with status
- **Settings**: Configure ICC profile and watch directory via file/folder dialogs
- **Open Log File**: Quickly open the log file in Notepad
- **Quit**: Exit the application

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
├── main.go                 # Application entry point
├── internal/
│   ├── logger/            # Logging functionality
│   ├── processor/         # Image processing and color profile application
│   ├── watcher/           # Directory watching
│   └── tray/              # System tray icon
├── .github/
│   └── workflows/
│       └── build.yml      # GitHub Actions workflow
└── README.md
```

## License

This project is open source and available under the MIT License.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Troubleshooting

### Application won't start
- For batch/single file modes, ensure you've specified a valid color profile with `-profile`
- For watch mode, you can configure the profile via the Settings menu after starting
- Check that the profile file exists and is readable

### Images not being processed
- Verify the watch directory is configured (via Settings or `-watch` flag)
- Ensure the watch directory exists and is accessible
- Check the log file for error messages
- Ensure images are in TIFF or JPEG format
- Verify that a color profile is configured

### System tray icon not appearing
- The system tray appears when running without arguments or in watch mode
- On Windows, check if the system tray is enabled

### Settings not saving
- Check that the application has write permissions in its directory
- The configuration is saved in `scanner_config.json`

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/adjust-scans/scanner).
