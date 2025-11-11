package singleton

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const mutexName = "Global\\Scanner_SingleInstance_Mutex"

var (
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	procCreateMutex = kernel32.NewProc("CreateMutexW")
	mutexHandle     windows.Handle
)

// TryLock attempts to acquire the singleton lock
// Returns true if successful, false if another instance is already running
func TryLock() (bool, error) {
	mutexNamePtr, err := syscall.UTF16PtrFromString(mutexName)
	if err != nil {
		return false, fmt.Errorf("failed to convert mutex name: %w", err)
	}

	ret, _, err := procCreateMutex.Call(
		0,
		0,
		uintptr(unsafe.Pointer(mutexNamePtr)),
	)

	if ret == 0 {
		return false, fmt.Errorf("failed to create mutex: %w", err)
	}

	mutexHandle = windows.Handle(ret)

	// Check if the mutex already existed
	if err == windows.ERROR_ALREADY_EXISTS {
		// Another instance is running
		windows.CloseHandle(mutexHandle)
		return false, nil
	}

	return true, nil
}

// Unlock releases the singleton lock
func Unlock() {
	if mutexHandle != 0 {
		windows.CloseHandle(mutexHandle)
		mutexHandle = 0
	}
}
