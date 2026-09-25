//go:build !windows

package main

func collectWindowsInfo(rep *Report) {
	// no-op on non-Windows platforms
}