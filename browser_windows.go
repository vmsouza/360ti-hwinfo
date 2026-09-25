//go:build windows

package main

import (
	"os/exec"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modUser32    = windows.NewLazySystemDLL("user32.dll")
	procMsgBox   = modUser32.NewProc("MessageBoxW")
)

// OpenInBrowser opens the generated report in the default Windows browser.
func OpenInBrowser(path string) {
	_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
}

// showFatalError shows a message box when the app cannot proceed.
func showFatalError(msg string) {
	title, _ := windows.UTF16PtrFromString("360ti HWiNFO")
	text, _ := windows.UTF16PtrFromString(msg)
	// MB_OK | MB_ICONERROR
	_, _, _ = procMsgBox.Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(title)), 0x10)
}