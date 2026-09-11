package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ArchiveItem struct {
	FullPath string
	Dir      string
	FileName string
	BaseName string
	Format   string
	Size     int64
	ModTime  time.Time
}

var supportedExtensions = []string{
	".zip",
	".rar",
	".7z",
	".tar",
	".tar.gz",
	".tgz",
	".tar.bz2",
	".tbz2",
	".tar.xz",
	".txz",
	".gz",
	".bz2",
	".xz",
}

func detectFormat(fileName string) (string, bool) {
	lower := strings.ToLower(fileName)
	switch {
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		return "TAR.GZ", true
	case strings.HasSuffix(lower, ".tar.bz2") || strings.HasSuffix(lower, ".tbz2"):
		return "TAR.BZ2", true
	case strings.HasSuffix(lower, ".tar.xz") || strings.HasSuffix(lower, ".txz"):
		return "TAR.XZ", true
	case strings.HasSuffix(lower, ".tar"):
		return "TAR", true
	case strings.HasSuffix(lower, ".zip"):
		return "ZIP", true
	case strings.HasSuffix(lower, ".rar"):
		return "RAR", true
	case strings.HasSuffix(lower, ".7z"):
		return "7Z", true
	case strings.HasSuffix(lower, ".gz"):
		return "GZ", true
	case strings.HasSuffix(lower, ".bz2"):
		return "BZ2", true
	case strings.HasSuffix(lower, ".xz"):
		return "XZ", true
	default:
		return "", false
	}
}

func getArchiveBaseName(filePath string) string {
	normalized := strings.ReplaceAll(filePath, "\\", "/")
	base := filepath.Base(normalized)
	lower := strings.ToLower(base)
	switch {
	case strings.HasSuffix(lower, ".tar.gz"):
		return base[:len(base)-7]
	case strings.HasSuffix(lower, ".tar.bz2"):
		return base[:len(base)-8]
	case strings.HasSuffix(lower, ".tar.xz"):
		return base[:len(base)-7]
	case strings.HasSuffix(lower, ".tgz"):
		return base[:len(base)-4]
	case strings.HasSuffix(lower, ".tbz2"):
		return base[:len(base)-5]
	case strings.HasSuffix(lower, ".txz"):
		return base[:len(base)-4]
	default:
		ext := filepath.Ext(base)
		if ext != "" {
			return base[:len(base)-len(ext)]
		}
		return base
	}
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetCandidateDirectories collects all plausible download and working directories
func GetCandidateDirectories(customDir string) []string {
	var dirs []string
	seen := make(map[string]bool)

	addDir := func(d string) {
		if strings.TrimSpace(d) == "" {
			return
		}
		clean := filepath.Clean(d)
		abs, err := filepath.Abs(clean)
		if err != nil {
			abs = clean
		}
		key := strings.ToLower(abs)
		if seen[key] {
			return
		}
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			seen[key] = true
			dirs = append(dirs, abs)
		}
	}

	// 1. Custom directory if provided
	if customDir != "" {
		addDir(customDir)
	}

	// 2. Current working directory
	if cwd, err := os.Getwd(); err == nil {
		addDir(cwd)
	}

	// 3. Executable's own directory
	if exePath, err := os.Executable(); err == nil {
		addDir(filepath.Dir(exePath))
	}

	// 4. USERPROFILE Downloads & Turkish İndirilenler
	userProfile := os.Getenv("USERPROFILE")
	if userProfile != "" {
		addDir(filepath.Join(userProfile, "Downloads"))
		addDir(filepath.Join(userProfile, "İndirilenler"))
		addDir(filepath.Join(userProfile, "Desktop"))
		addDir(filepath.Join(userProfile, "Masaüstü"))
	}

	// 5. System specific directories (systemprofile, all users' Downloads, registry, etc.)
	sysDirs := getSystemSpecificDirs()
	for _, d := range sysDirs {
		addDir(d)
	}

	return dirs
}

// ScanArchives searches for archives across given directories
func ScanArchives(dirs []string) ([]ArchiveItem, error) {
	var results []ArchiveItem
	seenPaths := make(map[string]bool)

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // skip inaccessible directories
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			format, isArchive := detectFormat(name)
			if !isArchive {
				continue
			}

			fullPath := filepath.Join(dir, name)
			canonical := strings.ToLower(filepath.Clean(fullPath))
			if seenPaths[canonical] {
				continue
			}
			seenPaths[canonical] = true

			info, err := entry.Info()
			if err != nil {
				continue
			}

			results = append(results, ArchiveItem{
				FullPath: fullPath,
				Dir:      dir,
				FileName: name,
				BaseName: getArchiveBaseName(name),
				Format:   format,
				Size:     info.Size(),
				ModTime:  info.ModTime(),
			})
		}
	}

	// Sort by modification time (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].ModTime.After(results[j].ModTime)
	})

	return results, nil
}

// MatchArchive finds an archive from input: index, filename, basename, or path
func MatchArchive(input string, archives []ArchiveItem, candidateDirs []string) (*ArchiveItem, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, fmt.Errorf("boş girdi")
	}

	// 1. Check if input is a 1-based index (e.g. "1", "2")
	if idx, err := strconv.Atoi(trimmed); err == nil {
		if idx >= 1 && idx <= len(archives) {
			return &archives[idx-1], nil
		}
		return nil, fmt.Errorf("geçersiz numara: %d (1-%d arasında olmalıdır)", idx, len(archives))
	}

	// 2. Check direct match by exact full path or relative path on disk
	if info, err := os.Stat(trimmed); err == nil && !info.IsDir() {
		absPath, _ := filepath.Abs(trimmed)
		format, isArchive := detectFormat(info.Name())
		if !isArchive {
			return nil, fmt.Errorf("desteklenmeyen dosya formatı: %s", filepath.Ext(info.Name()))
		}
		return &ArchiveItem{
			FullPath: absPath,
			Dir:      filepath.Dir(absPath),
			FileName: info.Name(),
			BaseName: getArchiveBaseName(info.Name()),
			Format:   format,
			Size:     info.Size(),
			ModTime:  info.ModTime(),
		}, nil
	}

	lowerInput := strings.ToLower(trimmed)

	// 3. Search in archives list by exact FileName (case-insensitive)
	for _, item := range archives {
		if strings.ToLower(item.FileName) == lowerInput {
			return &item, nil
		}
	}

	// 4. Search in archives list by BaseName (e.g. "a" matches "a.rar")
	var baseMatches []ArchiveItem
	for _, item := range archives {
		if strings.ToLower(item.BaseName) == lowerInput {
			baseMatches = append(baseMatches, item)
		}
	}
	if len(baseMatches) == 1 {
		return &baseMatches[0], nil
	} else if len(baseMatches) > 1 {
		return nil, fmt.Errorf("'%s' ile eşleşen birden fazla arşiv var. Lütfen uzantısıyla birlikte yazın (örn: %s)", trimmed, baseMatches[0].FileName)
	}

	// 5. Search in candidate directories if file exists on disk
	for _, dir := range candidateDirs {
		candidateFile := filepath.Join(dir, trimmed)
		if info, err := os.Stat(candidateFile); err == nil && !info.IsDir() {
			absPath, _ := filepath.Abs(candidateFile)
			format, isArchive := detectFormat(info.Name())
			if !isArchive {
				return nil, fmt.Errorf("desteklenmeyen arşiv formatı: %s", candidateFile)
			}
			return &ArchiveItem{
				FullPath: absPath,
				Dir:      dir,
				FileName: info.Name(),
				BaseName: getArchiveBaseName(info.Name()),
				Format:   format,
				Size:     info.Size(),
				ModTime:  info.ModTime(),
			}, nil
		}
	}

	return nil, fmt.Errorf("'%s' adında bir arşiv bulunamadı", trimmed)
}
