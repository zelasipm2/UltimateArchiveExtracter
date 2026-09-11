//go:build windows

package main

import (
	"golang.org/x/sys/windows"
)

func initConsole() {
	// Set console code page to UTF-8 (65001)
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	setConsoleOutputCP := kernel32.NewProc("SetConsoleOutputCP")
	setConsoleCP := kernel32.NewProc("SetConsoleCP")

	if setConsoleOutputCP.Find() == nil {
		setConsoleOutputCP.Call(65001)
	}
	if setConsoleCP.Find() == nil {
		setConsoleCP.Call(65001)
	}

	// Enable Virtual Terminal Processing for ANSI colors if supported
	hOut, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err == nil && hOut != windows.InvalidHandle {
		var mode uint32
		if err := windows.GetConsoleMode(hOut, &mode); err == nil {
			mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
			_ = windows.SetConsoleMode(hOut, mode)
		}
	}
}
