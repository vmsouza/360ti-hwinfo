//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
)

// OpenInBrowser opens the generated report in the default browser (non-Windows).
func OpenInBrowser(path string) {
	cmd := exec.Command("xdg-open", path)
	if err := cmd.Start(); err != nil {
		fmt.Println("could not open browser:", err)
	}
}

// showFatalError prints the message to the console (non-Windows).
func showFatalError(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}