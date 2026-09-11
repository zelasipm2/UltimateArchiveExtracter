package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// checkZipSlip ensures target path does not escape destDir
func checkZipSlip(destDir, entryName string) (string, error) {
	// Normalize slashes
	cleanEntry := filepath.Clean(entryName)

	// Avoid leading slashes or volume names
	if filepath.IsAbs(cleanEntry) || strings.HasPrefix(cleanEntry, "..") {
		return "", fmt.Errorf("güvenlik uyarısı (Zip Slip tespit edildi): %s", entryName)
	}

	target := filepath.Join(destDir, cleanEntry)
	cleanTarget := filepath.Clean(target)
	cleanDest := filepath.Clean(destDir)

	if !strings.HasPrefix(cleanTarget, cleanDest+string(filepath.Separator)) && cleanTarget != cleanDest {
		return "", fmt.Errorf("güvenlik uyarısı (klasör dışına taşma tespit edildi): %s", entryName)
	}

	return cleanTarget, nil
}

// ExtractArchive unpacks an archive into a folder named after the archive
func ExtractArchive(item *ArchiveItem, customOutputDir string, password string, verbose bool) (string, int, error) {
	// Determine output directory: "o dosyayı klasör ismiyle çıkartacaksın"
	var destDir string
	if customOutputDir != "" {
		destDir = filepath.Join(customOutputDir, item.BaseName)
	} else {
		destDir = filepath.Join(item.Dir, item.BaseName)
	}

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", 0, fmt.Errorf("hedef klasör oluşturulamadı (%s): %w", destDir, err)
	}

	fmt.Printf("[*] Arşiv Açılıyor : %s\n", item.FileName)
	fmt.Printf("[*] Format         : %s (%s)\n", item.Format, formatSize(item.Size))
	fmt.Printf("[*] Hedef Klasör   : %s\n", destDir)
	fmt.Println("----------------------------------------------------------------")

	var count int
	var err error

	switch item.Format {
	case "ZIP":
		count, err = extractZip(item.FullPath, destDir, password, verbose)
	case "7Z":
		count, err = extract7z(item.FullPath, destDir, password, verbose)
	case "RAR":
		count, err = extractRar(item.FullPath, destDir, password, verbose)
	case "TAR", "TAR.GZ", "TAR.BZ2", "TAR.XZ":
		count, err = extractTar(item.FullPath, destDir, item.Format, verbose)
	case "GZ", "BZ2", "XZ":
		count, err = extractSingleCompressed(item.FullPath, destDir, item.Format, item.BaseName)
	default:
		return "", 0, fmt.Errorf("desteklenmeyen arşiv biçimi: %s", item.Format)
	}

	if err != nil {
		return destDir, count, err
	}

	return destDir, count, nil
}
