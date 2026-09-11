//go:build windows

package main

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func getSystemSpecificDirs() []string {
	var dirs []string

	// 1. Check User Shell Folders in Registry for current user
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Explorer\User Shell Folders`, registry.QUERY_VALUE)
	if err == nil {
		defer k.Close()
		// FOLDERID_Downloads: {374DE290-123F-4565-9164-39C4925E467B}
		if val, _, err := k.GetStringValue("{374DE290-123F-4565-9164-39C4925E467B}"); err == nil && val != "" {
			dirs = append(dirs, os.ExpandEnv(val))
		}
		// Shell Folders "Downloads"
		if val, _, err := k.GetStringValue("Downloads"); err == nil && val != "" {
			dirs = append(dirs, os.ExpandEnv(val))
		}
	}

	// 2. System Root / WinDir (e.g. C:\Windows)
	sysRoot := os.Getenv("SystemRoot")
	if sysRoot == "" {
		sysRoot = os.Getenv("WINDIR")
	}
	if sysRoot == "" {
		sysRoot = `C:\Windows`
	}

	// systemprofile directories (crucial when running as LocalSystem / system32\systemprofile)
	systemProfileDir := filepath.Join(sysRoot, "System32", "config", "systemprofile")
	dirs = append(dirs,
		filepath.Join(systemProfileDir, "Downloads"),
		filepath.Join(systemProfileDir, "İndirilenler"),
		filepath.Join(systemProfileDir, "Desktop"),
		filepath.Join(systemProfileDir, "Masaüstü"),
		systemProfileDir,
		filepath.Join(sysRoot, "System32"),
		filepath.Join(sysRoot, "Temp"),
	)

	// 3. SystemDrive Users directory (e.g. C:\Users)
	sysDrive := os.Getenv("SystemDrive")
	if sysDrive == "" {
		sysDrive = "C:"
	}
	usersBase := filepath.Join(sysDrive, "Users")
	if entries, err := os.ReadDir(usersBase); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := strings.ToLower(entry.Name())
			if name == "public" || name == "default" || name == "default user" || name == "all users" {
				continue
			}
			userDir := filepath.Join(usersBase, entry.Name())
			dirs = append(dirs,
				filepath.Join(userDir, "Downloads"),
				filepath.Join(userDir, "İndirilenler"),
				filepath.Join(userDir, "Desktop"),
				filepath.Join(userDir, "Masaüstü"),
			)
		}
	}

	return dirs
}
