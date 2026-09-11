//go:build !windows

package main

import (
	"os"
	"path/filepath"
)

func getSystemSpecificDirs() []string {
	var dirs []string
	home := os.Getenv("HOME")
	if home != "" {
		dirs = append(dirs,
			filepath.Join(home, "Downloads"),
			filepath.Join(home, "Desktop"),
		)
	}
	dirs = append(dirs, "/tmp")
	return dirs
}
