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

### Watch Directory Mode

Monitor a directory for new images and process them automatically:

```bash
scanner.exe -watch "C:\Scans" -profile "C:\Profiles\sRGB.icc"
```

This will:
- Watch the `C:\Scans` directory for new TIFF/JPG files
- Apply the color profile from `sRGB.icc`
- Save corrected images to `C:\Scans\fixed\`
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
| `-watch` | Directory to watch for new images | * |
| `-process-dir` | Process all images in directory and exit | * |
| `-process-file` | Process a specific file and exit | * |
| `-profile` | Path to ICC color profile file | Yes |
| `-output` | Output subdirectory name (default: "fixed") | No |
| `-log` | Log file path (default: "scanner.log") | No |

*One of `-watch`, `-process-dir`, or `-process-file` must be specified.

## System Tray Features

When running in watch mode, the system tray icon provides:

- **Recent Images**: View the last 10 processed images with status
- **Open Log File**: Quickly open the log file in Notepad
- **Quit**: Exit the application

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
- Ensure you've specified a valid color profile with `-profile`
- Check that the profile file exists and is readable

### Images not being processed
- Verify the watch directory exists and is accessible
- Check the log file for error messages
- Ensure images are in TIFF or JPEG format

### System tray icon not appearing
- This feature only works in watch mode (`-watch` flag)
- On Windows, check if the system tray is enabled

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/adjust-scans/scanner).
