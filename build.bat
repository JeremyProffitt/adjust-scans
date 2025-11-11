@echo off
echo Building scanner application...
go build -v -ldflags="-H=windowsgui" -o scanner.exe .
if %errorlevel% equ 0 (
    echo.
    echo Build successful: scanner.exe
) else (
    echo.
    echo Build failed!
    exit /b 1
)
