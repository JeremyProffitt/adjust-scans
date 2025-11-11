# Scanner - Architecture Documentation

This document provides a comprehensive technical overview of the Scanner application architecture, design decisions, and implementation details for developers.

## Table of Contents
- [Overview](#overview)
- [Architecture](#architecture)
- [Module Descriptions](#module-descriptions)
- [Data Flow](#data-flow)
- [Technology Stack](#technology-stack)
- [Design Patterns](#design-patterns)
- [Platform Considerations](#platform-considerations)
- [Build System](#build-system)
- [Testing Strategy](#testing-strategy)
- [Extension Points](#extension-points)

## Overview

Scanner is a Windows-native GUI application built in Go that applies ICC color profiles to scanned images. It follows a modular architecture with clear separation of concerns, making it maintainable and extensible.

### Key Design Goals
1. **User-Friendly**: Minimal UI complexity - runs in system tray
2. **Reliable**: Single instance, robust error handling, comprehensive logging
3. **Performant**: Efficient image processing, minimal resource usage
4. **Platform-Integrated**: Native Windows APIs for tray, startup, and file operations
5. **Maintainable**: Clean module boundaries, testable code

## Architecture

### High-Level Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         main.go                              │
│  - Entry point                                               │
│  - Singleton lock acquisition                                │
│  - Mode selection (tray/watch/batch/single)                  │
│  - Dependency injection                                      │
└────────────┬────────────────────────────────────────────────┘
             │
             ├───────────────────┬────────────────────┬─────────────┐
             │                   │                    │             │
             ▼                   ▼                    ▼             ▼
      ┌──────────┐        ┌──────────┐        ┌──────────┐   ┌──────────┐
      │   Tray   │        │ Watcher  │        │Processor │   │  Logger  │
      │  Module  │        │  Module  │        │  Module  │   │  Module  │
      └────┬─────┘        └────┬─────┘        └────┬─────┘   └────┬─────┘
           │                   │                    │              │
           │                   │                    │              │
           ├───────────────────┴────────────────────┴──────────────┘
           │
           ├──────────┬──────────┬─────────────┬──────────┐
           │          │          │             │          │
           ▼          ▼          ▼             ▼          ▼
      ┌────────┐ ┌────────┐ ┌────────┐  ┌────────┐ ┌─────────┐
      │ Config │ │Startup │ │  Icon  │  │Dialogs │ │Singleton│
      └────────┘ └────────┘ └────────┘  └────────┘ └─────────┘
```

### Module Interaction Flow

1. **Startup**: `main.go` → `singleton` → `logger` → `config`
2. **Tray Mode**: `tray` → `processor` + `config` + `startup`
3. **Watch Mode**: `watcher` → `processor`, with `tray` for UI
4. **Batch Mode**: `processor` directly (no tray)

## Module Descriptions

### main.go - Application Entry Point

**Location**: `/main.go`

**Responsibilities**:
- Parse command-line flags
- Acquire singleton lock (prevent multiple instances)
- Initialize logging subsystem
- Load/create configuration
- Determine operation mode (tray/watch/batch/single)
- Coordinate module initialization
- Manage application lifecycle

**Key Functions**:
- `main()`: Entry point, orchestrates startup
- `processSingleFile()`: Single-file processing mode
- `processDirectory()`: Batch processing mode
- `watchDirectory()`: Directory watching mode
- `trayMode()`: GUI-only tray mode

**Dependencies**: All modules (coordinates everything)

---

### internal/singleton - Single Instance Enforcement

**Location**: `/internal/singleton/`

**Files**:
- `singleton_windows.go`: Windows implementation using named mutexes
- `singleton_other.go`: Unix/Linux implementation using lock files

**Responsibilities**:
- Prevent multiple application instances
- Platform-specific implementation using build tags
- Clean resource management

**Key Functions**:
- `TryLock() (bool, error)`: Attempt to acquire singleton lock
- `Unlock()`: Release the lock

**Windows Implementation**:
```go
// Uses Win32 CreateMutex API
// Mutex name: "Global\\Scanner_SingleInstance_Mutex"
// Returns false if ERROR_ALREADY_EXISTS
```

**Design Rationale**:
- Windows: Named mutex is the standard approach and survives crashes
- Unix: Lock file in temp directory (for potential cross-platform support)

---

### internal/logger - Logging Subsystem

**Location**: `/internal/logger/`

**Responsibilities**:
- Centralized logging to file
- Thread-safe log operations
- Structured log levels (INFO, ERROR, etc.)
- Automatic file creation and management

**Key Types**:
```go
type Logger struct {
    file   *os.File
    logger *log.Logger
    mu     sync.Mutex
}
```

**Key Functions**:
- `New(filePath string) (*Logger, error)`: Create logger instance
- `Info(format string, args ...interface{})`: Info-level logging
- `Infof(format string, args ...interface{})`: Formatted info logging
- `Error(format string, args ...interface{})`: Error-level logging
- `Errorf(format string, args ...interface{})`: Formatted error logging
- `Close()`: Clean shutdown

**Thread Safety**: Uses mutex for concurrent write protection

**Output Format**:
```
2025/11/11 04:35:23 [INFO] Scanner application started
2025/11/11 04:35:24 [ERROR] Failed to process image.tiff: profile not found
```

---

### internal/config - Configuration Management

**Location**: `/internal/config/`

**Responsibilities**:
- Load/save JSON configuration file
- Provide thread-safe access to settings
- Default value management
- Validation

**Key Types**:
```go
type Config struct {
    ProfilePath string `json:"profile_path"`
    WatchDir    string `json:"watch_dir"`
    OutputDir   string `json:"output_dir"`
    mu          sync.RWMutex
}
```

**Configuration File**: `scanner_config.json`
```json
{
  "profile_path": "C:\\path\\to\\profile.icc",
  "watch_dir": "C:\\path\\to\\watch",
  "output_dir": "fixed"
}
```

**Key Functions**:
- `Load() (*Config, error)`: Load or create default config
- `GetProfilePath() string`: Thread-safe read
- `SetProfilePath(path string) error`: Thread-safe write + persist
- `GetWatchDir() string`: Thread-safe read
- `SetWatchDir(path string) error`: Thread-safe write + persist
- `GetOutputDir() string`: Thread-safe read
- `SetOutputDir(name string) error`: Thread-safe write + persist
- `save() error`: Persist to disk

**Thread Safety**: RWMutex for concurrent read/write protection

---

### internal/processor - Image Processing Engine

**Location**: `/internal/processor/`

**Responsibilities**:
- Apply ICC color profiles to images
- Support TIFF and JPEG formats
- Track recently processed images
- Provide processing history

**Key Types**:
```go
type Processor struct {
    profilePath   string
    recentImages  []ProcessedImage
    mu            sync.RWMutex
    log           *logger.Logger
}

type ProcessedImage struct {
    FileName       string
    ProcessedTime  time.Time
    Success        bool
    Error          string
}
```

**Key Functions**:
- `New(profilePath string, log *logger.Logger) *Processor`: Constructor
- `ProcessImage(inputPath, outputDir string) error`: Main processing function
- `UpdateProfile(profilePath string) error`: Hot-swap color profile
- `GetRecentImages() []ProcessedImage`: Retrieve processing history (last 10)

**Processing Pipeline**:
1. Decode input image (TIFF/JPEG)
2. Load ICC profile
3. Apply color transformation
4. Encode output image
5. Save to output directory
6. Update recent images list

**Dependencies**:
- `github.com/disintegration/imaging`: Image decoding/encoding
- `golang.org/x/image/tiff`: TIFF format support
- ICC profile parsing (custom implementation)

**Error Handling**:
- Graceful failure with detailed error messages
- Continues processing other files in batch mode
- Tracks failures in recent images list

---

### internal/watcher - File System Monitoring

**Location**: `/internal/watcher/`

**Responsibilities**:
- Monitor directory for new files
- Debounce file system events
- Trigger image processing automatically
- Filter by file extension

**Key Types**:
```go
type Watcher struct {
    watchDir   string
    outputDir  string
    processor  *processor.Processor
    watcher    *fsnotify.Watcher
    log        *logger.Logger
    stopChan   chan struct{}
}
```

**Key Functions**:
- `New(watchDir, outputDir string, proc *Processor, log *Logger) (*Watcher, error)`
- `Start() error`: Begin watching directory
- `Stop()`: Clean shutdown
- `handleEvent(event fsnotify.Event)`: Process file system events

**Event Handling**:
```go
// Monitors for:
- fsnotify.Create: New files created
- fsnotify.Write: Files modified (debounced)

// Filters:
- Ignores directories
- Only processes: .tiff, .tif, .jpg, .jpeg
- Skips files in output directory
```

**Dependencies**:
- `github.com/fsnotify/fsnotify`: Cross-platform file system notifications

**Design Considerations**:
- Debouncing prevents duplicate processing during file writes
- Separate goroutine for event processing
- Clean shutdown via stop channel

---

### internal/tray - System Tray GUI

**Location**: `/internal/tray/`

**Files**:
- `tray.go`: Main tray implementation
- `scanner_icon.ico`: Embedded icon (multi-resolution)

**Responsibilities**:
- System tray icon management
- Context menu creation and event handling
- File/folder dialogs
- Integration with all other modules

**Key Types**:
```go
type Tray struct {
    log            *logger.Logger
    processor      *processor.Processor
    config         *config.Config
    logFilePath    string
    onConfigChange func()
}
```

**Key Functions**:
- `New(...) (*Tray, error)`: Constructor with dependency injection
- `Run()`: Start tray (blocks until quit)
- `onReady()`: Initialize menu (called by systray library)
- `setProfile()`: Profile selection dialog + config update
- `setWatchDirectory()`: Directory selection dialog + config update
- `openWatchDirectory()`: Open folder in Explorer
- `processFile()`: Single file processing + auto-open result
- `processDirectory()`: Batch processing + auto-open output folder
- `toggleStartAutomatically()`: Enable/disable auto-start
- `openLogFile()`: Open log in Notepad
- `openFile(path string)`: OS-agnostic file opening
- `openFolder(path string)`: OS-agnostic folder opening

**Menu Structure**:
```
Scanner (Icon)
├── Set Profile
├── Set Watch Directory
├── Open Watch Directory
├── ────────────────────
├── Process File
├── Process Directory
├── ────────────────────
├── Start Automatically ☑
├── ────────────────────
├── Open Log File
├── ────────────────────
└── Quit
```

**Icon Embedding**:
```go
//go:embed scanner_icon.ico
var iconData []byte

func (t *Tray) onReady() {
    systray.SetIcon(iconData)
    // ...
}
```

**Dependencies**:
- `github.com/getlantern/systray`: System tray library
- `github.com/sqweek/dialog`: Native file/folder dialogs

**Design Rationale**:
- Embedded icon eliminates external file dependency
- Callback-based menu system for clean event handling
- Dependency injection makes testing possible
- Auto-open features improve UX (immediate feedback)

---

### internal/startup - Windows Auto-Start Management

**Location**: `/internal/startup/`

**Files**:
- `startup_windows.go`: Windows Registry implementation
- `startup_other.go`: Stub for non-Windows platforms

**Responsibilities**:
- Add/remove application from Windows startup
- Check current auto-start status
- Windows Registry manipulation

**Key Functions**:
- `IsEnabled() bool`: Check if auto-start is enabled
- `Enable() error`: Add to Windows startup
- `Disable() error`: Remove from Windows startup

**Windows Implementation**:
```go
const registryPath = `Software\Microsoft\Windows\CurrentVersion\Run`
const appName = "Scanner"

// Writes executable path to:
// HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run
```

**Registry Value**:
```
Key: Scanner
Value: C:\path\to\scanner.exe
```

**Dependencies**:
- `golang.org/x/sys/windows/registry`: Windows Registry access

**Error Handling**:
- Returns errors for permission issues
- Gracefully handles missing registry keys

---

## Data Flow

### Tray Mode with Manual Processing

```
User clicks "Process File"
    │
    ▼
Tray shows file picker dialog
    │
    ▼
User selects image.tiff
    │
    ▼
Tray.processFile() called
    │
    ▼
Processor.ProcessImage("image.tiff", "output/")
    │
    ├─> Load image
    ├─> Load ICC profile
    ├─> Apply color transformation
    ├─> Save to output/image.tiff
    └─> Update recent images
    │
    ▼
Tray.openFile("output/image.tiff")
    │
    ▼
OS opens image in default viewer
```

### Watch Mode with Auto-Processing

```
Application starts in watch mode
    │
    ▼
Watcher monitors C:\Scans
    │
    ▼
New file: C:\Scans\image.tiff created
    │
    ▼
Watcher.handleEvent(CREATE)
    │
    ▼
Filter check: .tiff extension? Yes
    │
    ▼
Processor.ProcessImage("C:\Scans\image.tiff", "C:\Scans\fixed\")
    │
    ├─> Process image
    └─> Log result
    │
    ▼
Image saved to C:\Scans\fixed\image.tiff
    │
    ▼
Tray remains available for user interaction
```

### Configuration Update Flow

```
User clicks "Set Profile"
    │
    ▼
Tray shows file picker (.icc files)
    │
    ▼
User selects profile.icc
    │
    ▼
Config.SetProfilePath("C:\profiles\profile.icc")
    │
    ├─> Update in-memory config (with mutex)
    └─> Save to scanner_config.json
    │
    ▼
Processor.UpdateProfile("C:\profiles\profile.icc")
    │
    └─> Reload ICC profile for future processing
    │
    ▼
onConfigChange() callback fires
    │
    └─> Logs configuration change
```

## Technology Stack

### Core Language
- **Go 1.25.0**: Modern, type-safe, compiled language
  - Cross-compilation support (though Windows-focused)
  - Strong standard library
  - Excellent concurrency primitives
  - Fast compilation and execution

### Key Dependencies

#### UI & System Integration
- **github.com/getlantern/systray** (v1.2.2)
  - Cross-platform system tray library
  - Callback-based event system
  - Icon and tooltip support

- **github.com/sqweek/dialog** (v0.0.0-20240226140203-065105509627)
  - Native file/folder picker dialogs
  - Windows, macOS, Linux support
  - Clean API

#### Image Processing
- **github.com/disintegration/imaging** (v1.6.2)
  - Image manipulation and transformation
  - Format conversion
  - Resizing, color adjustment

- **golang.org/x/image** (v0.32.0)
  - Extended image format support
  - TIFF codec
  - Color space management

#### File System Monitoring
- **github.com/fsnotify/fsnotify** (v1.9.0)
  - Cross-platform file system notifications
  - Event-based API (Create, Write, Remove, Rename)
  - Production-ready, widely used

#### Windows Platform
- **golang.org/x/sys/windows** (v0.13.0)
  - Windows API bindings
  - Registry access
  - System calls

### Build Tools
- **Go toolchain**: Compilation
- **GitHub Actions**: CI/CD
- **build.bat**: Convenience script

## Design Patterns

### 1. Dependency Injection

Used extensively for testability and flexibility:

```go
// Example: Tray requires logger, processor, config
func New(log *logger.Logger, proc *processor.Processor,
         cfg *config.Config, logFilePath string,
         onConfigChange func()) (*Tray, error)
```

**Benefits**:
- Testable (can inject mocks)
- Flexible (swap implementations)
- Clear dependencies

### 2. Singleton Pattern

Enforced at application level via `singleton` package:

```go
locked, err := singleton.TryLock()
if !locked {
    // Another instance running
    os.Exit(1)
}
defer singleton.Unlock()
```

**Benefits**:
- Prevents conflicts (file locks, tray icon)
- Resource management
- User experience (clear error)

### 3. Observer Pattern

Configuration change callback:

```go
onConfigChange := func() {
    log.Info("Configuration changed")
}
tray := tray.New(..., onConfigChange)
```

**Benefits**:
- Loose coupling
- Event-driven updates

### 4. Strategy Pattern

Platform-specific implementations using build tags:

```go
// startup_windows.go
//go:build windows

// startup_other.go
//go:build !windows
```

**Benefits**:
- Clean platform abstraction
- Compile-time selection
- No runtime overhead

### 5. Repository Pattern

Configuration management:

```go
type Config struct {
    // Storage abstraction
    ProfilePath string
    WatchDir    string
    // ...
}

func Load() (*Config, error)  // Load from storage
func (c *Config) save() error // Persist to storage
```

**Benefits**:
- Abstracts storage mechanism
- Easy to test
- Easy to migrate storage backend

## Platform Considerations

### Windows-Specific Features

#### System Tray
- Uses `getlantern/systray` library
- Native Windows tray integration
- Context menu via Win32 API

#### Auto-Start
- Windows Registry modification
- `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run`
- No elevation required (per-user)

#### Singleton Lock
- Named mutex via `CreateMutexW` Win32 API
- Global namespace for system-wide uniqueness
- Survives crashes (OS cleans up)

#### File Operations
- Windows path format (backslashes)
- `explorer` for opening folders
- `cmd /c start` for opening files
- `notepad.exe` for log files

#### GUI Mode Build Flag
```bash
go build -ldflags="-H=windowsgui"
```
- Suppresses console window
- Changes executable subsystem to "Windows GUI"
- Allows background execution

### Cross-Platform Considerations

While Windows-focused, platform abstraction exists:

```go
// Cross-platform file opening
switch runtime.GOOS {
case "windows":
    cmd = exec.Command("cmd", "/c", "start", "", filePath)
case "darwin":
    cmd = exec.Command("open", filePath)
default:
    cmd = exec.Command("xdg-open", filePath)
}
```

**Portability**: Core logic (image processing, watching) works on Unix/Linux; GUI features are Windows-optimized.

## Build System

### Manual Build

```bash
go build -v -ldflags="-H=windowsgui" -o scanner.exe .
```

**Flags Explained**:
- `-v`: Verbose output (shows compiled packages)
- `-ldflags="-H=windowsgui"`: Windows GUI subsystem (no console)
- `-o scanner.exe`: Output filename

### Convenience Script

`build.bat`:
```batch
@echo off
echo Building scanner application...
go build -v -ldflags="-H=windowsgui" -o scanner.exe .
if %errorlevel% equ 0 (
    echo Build successful: scanner.exe
) else (
    echo Build failed!
    exit /b 1
)
```

### CI/CD Pipeline

`.github/workflows/build.yml`:
- Triggered on push to main/master
- Runs on `windows-latest`
- Steps:
  1. Checkout code
  2. Setup Go 1.21
  3. Download dependencies
  4. Run tests
  5. Build with GUI flag
  6. Upload artifact
  7. (On release) Upload to GitHub Releases

**Artifacts**: `scanner-windows-amd64.exe`

## Testing Strategy

### Unit Tests

Located in `*_test.go` files alongside source:

- `internal/logger/logger_test.go`
- `internal/processor/processor_test.go`
- `internal/watcher/watcher_test.go`

**Run All Tests**:
```bash
go test ./...
```

**Coverage**:
```bash
go test -cover ./...
```

### Test Coverage Goals
- Logger: 100% (critical component)
- Processor: >80% (core functionality)
- Watcher: >80% (complex async behavior)
- Config: >70% (simple CRUD)

### Testing Challenges

**System Tray**: Not easily testable (requires user interaction)
- Focus on logic outside `systray.Run()`
- Test individual handler functions

**File System**: Use temporary directories
```go
tempDir := t.TempDir() // Auto-cleanup
```

**Windows Registry**: Mock or skip on non-Windows
```go
if runtime.GOOS != "windows" {
    t.Skip("Windows-only test")
}
```

### Manual Testing Checklist

- [ ] System tray icon appears
- [ ] All menu items clickable
- [ ] Set Profile updates config
- [ ] Set Watch Directory updates config
- [ ] Process File opens result
- [ ] Process Directory opens output folder
- [ ] Start Automatically adds registry entry
- [ ] Second instance is blocked
- [ ] Log file is written
- [ ] Auto-start works after reboot

## Extension Points

### Adding New Image Formats

1. Update file extension filter in `watcher/watcher.go`:
```go
ext := filepath.Ext(filePath)
if ext == ".tiff" || ext == ".tif" || ext == ".jpg" ||
   ext == ".jpeg" || ext == ".png" { // Add PNG
```

2. Update processor to handle new format:
```go
case ".png":
    img, err = imaging.Open(inputPath, imaging.AutoOrientation())
```

3. Update README documentation

### Adding New Menu Items

1. Add menu item in `tray.go`:
```go
mNewFeature := systray.AddMenuItem("New Feature", "Description")
```

2. Add handler in event loop:
```go
case <-mNewFeature.ClickedCh:
    t.handleNewFeature()
```

3. Implement handler function:
```go
func (t *Tray) handleNewFeature() {
    // Implementation
}
```

### Adding Configuration Options

1. Update `Config` struct in `internal/config/config.go`:
```go
type Config struct {
    ProfilePath string `json:"profile_path"`
    WatchDir    string `json:"watch_dir"`
    OutputDir   string `json:"output_dir"`
    NewOption   string `json:"new_option"` // Add here
}
```

2. Add getter/setter:
```go
func (c *Config) GetNewOption() string { /* ... */ }
func (c *Config) SetNewOption(value string) error { /* ... */ }
```

3. Update default values in `Load()`:
```go
return &Config{
    OutputDir: "fixed",
    NewOption: "default_value", // Add here
}, nil
```

### Adding Logging Levels

Current: INFO, ERROR

To add DEBUG, WARN:

1. Update `Logger` in `internal/logger/logger.go`:
```go
func (l *Logger) Debug(msg string) { /* ... */ }
func (l *Logger) Warn(msg string) { /* ... */ }
```

2. Add level filtering:
```go
type LogLevel int
const (
    DEBUG LogLevel = iota
    INFO
    WARN
    ERROR
)
```

### Cross-Platform GUI

To support macOS/Linux:

1. Test existing `systray` behavior on target platform
2. Update platform-specific code in `startup` and `singleton`
3. Update build scripts
4. Test file dialogs on target platform

---

## Code Organization Principles

### Package Structure
- **Flat `internal/` hierarchy**: Easy navigation
- **Single responsibility**: Each package has one job
- **No circular dependencies**: Clean dependency graph

### Naming Conventions
- **Packages**: Lowercase, single word (e.g., `logger`, `watcher`)
- **Exported types**: PascalCase (e.g., `Processor`, `Config`)
- **Unexported types**: camelCase (e.g., `recentImages`)
- **Interfaces**: Minimal, named with `-er` suffix (e.g., `Logger`)

### Error Handling
```go
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```
- Use `%w` for error wrapping (Go 1.13+)
- Return errors up the stack
- Log at top level only
- Provide context in error messages

### Concurrency
- **Mutexes**: Protect shared state (`Config`, `Processor.recentImages`)
- **Channels**: For signaling (`watcher.stopChan`)
- **Goroutines**: Event loops, background tasks
- **No naked goroutines**: Always manage lifecycle

---

## Performance Considerations

### Image Processing
- **Memory**: Large images can consume significant RAM
- **CPU**: Color transformation is CPU-intensive
- **I/O**: Disk reads/writes are slowest operations

**Optimizations**:
- Stream processing (don't load entire image into memory if possible)
- Limit concurrent processing (if batch mode)
- Use buffered I/O

### File Watching
- **Event storms**: Many files created simultaneously
- **Debouncing**: Wait for file to settle before processing

### Logging
- **Buffered writes**: Logger uses buffered I/O
- **Mutex contention**: Minimal (quick writes)

---

## Security Considerations

### Input Validation
- **File paths**: Validated before processing
- **ICC profiles**: Parsed safely (bounds checking)
- **Configuration**: JSON parsing with type safety

### Privileges
- **No elevation needed**: Runs as current user
- **Registry writes**: User hive only (HKCU)
- **File access**: Only directories user has access to

### Attack Surface
- **Minimal**: No network exposure
- **Local only**: All operations are local file system
- **No external input**: User selects files via dialogs

---

## Future Enhancements

### Potential Features
1. **Multiple profiles**: Support profile presets
2. **Batch rename**: Auto-rename based on date/sequence
3. **Preview**: Show before/after comparison
4. **OCR integration**: Automatically extract text
5. **Cloud sync**: Auto-upload processed images
6. **Notifications**: Toast notifications for processed files
7. **Statistics**: Track processing metrics
8. **Scheduled processing**: Process at specific times

### Architecture Improvements
1. **Plugin system**: Allow custom processors
2. **Database**: SQLite for configuration and history
3. **REST API**: Control via HTTP (for integration)
4. **Scriptable**: Lua/JavaScript for custom workflows
5. **Settings UI**: Full dialog instead of file pickers
6. **Updater**: Auto-update mechanism

---

## Contributing Guidelines

### Code Style
- Follow standard Go conventions
- Run `go fmt` before committing
- Use `go vet` to catch issues
- Keep functions small (<50 lines)
- Comment exported symbols

### Pull Request Process
1. Create feature branch
2. Write tests for new code
3. Update documentation
4. Ensure tests pass
5. Submit PR with clear description

### Testing Requirements
- All new features must have tests
- Maintain >70% coverage
- Include manual testing steps in PR description

---

## Debugging Tips

### Enable Verbose Logging
Modify `logger` to include DEBUG level for development.

### Check Log File
Always check `scanner.log` first:
```
2025/11/11 04:35:23 [INFO] Scanner application started
2025/11/11 04:35:24 [ERROR] Failed to load profile: file not found
```

### Attach Debugger
Use `dlv` (Delve) for Go debugging:
```bash
go install github.com/go-delve/delve/cmd/dlv@latest
dlv debug
```

### Build Without GUI Flag
For debugging with console output:
```bash
go build -o scanner_debug.exe .
./scanner_debug.exe
```

### Check Registry
Verify auto-start entry:
```batch
reg query "HKCU\Software\Microsoft\Windows\CurrentVersion\Run" /v Scanner
```

---

## License
This project is open source and available under the MIT License.

## Contact
For questions or contributions, please visit the [GitHub repository](https://github.com/adjust-scans/scanner).
