package main

import (
	"path/filepath"
	"testing"
	"time"
)

func TestDetectFormat(t *testing.T) {
	cases := map[string]string{
		"a.zip":          "ZIP",
		"data.rar":       "RAR",
		"archive.7z":     "7Z",
		"backup.tar":     "TAR",
		"file.tar.gz":    "TAR.GZ",
		"file.tgz":       "TAR.GZ",
		"data.tar.bz2":   "TAR.BZ2",
		"data.tbz2":      "TAR.BZ2",
		"source.tar.xz":  "TAR.XZ",
		"source.txz":     "TAR.XZ",
		"raw.gz":         "GZ",
		"raw.bz2":        "BZ2",
		"raw.xz":         "XZ",
		"UPPER.ZIP":      "ZIP",
		"Upper.Rar":      "RAR",
		"UPPER.7Z":       "7Z",
	}

	for filename, expected := range cases {
		format, ok := detectFormat(filename)
		if !ok || format != expected {
			t.Errorf("detectFormat(%q) = (%q, %v); expected (%q, true)", filename, format, ok, expected)
		}
	}

	// Unsupported
	_, ok := detectFormat("file.exe")
	if ok {
		t.Errorf("expected false for file.exe")
	}
}

func TestGetArchiveBaseName(t *testing.T) {
	cases := map[string]string{
		"a.rar":                   "a",
		"test.zip":                "test",
		"my.archive.7z":           "my.archive",
		"backup.tar.gz":           "backup",
		"package.tar.bz2":         "package",
		"source.tar.xz":           "source",
		"fast.tgz":                "fast",
		"compressed.tbz2":         "compressed",
		"kernel.txz":              "kernel",
		"/path/to/archive.rar":    "archive",
		"C:\\Users\\test\\a.zip":  "a",
	}

	for input, expected := range cases {
		got := getArchiveBaseName(input)
		if got != expected {
			t.Errorf("getArchiveBaseName(%q) = %q; expected %q", input, got, expected)
		}
	}
}

func TestCheckZipSlip(t *testing.T) {
	dest := "/tmp/safe_dir"

	// Valid entries
	valid := []string{
		"file.txt",
		"sub/folder/file.txt",
		"sub/./file.txt",
	}
	for _, v := range valid {
		target, err := checkZipSlip(dest, v)
		if err != nil {
			t.Errorf("expected %q to be valid, got err: %v", v, err)
		}
		if target == "" {
			t.Errorf("expected non-empty target for %q", v)
		}
	}

	// Malicious Zip Slip entries
	malicious := []string{
		"../evil.txt",
		"../../evil.txt",
		"/etc/passwd",
		"sub/../../evil.txt",
		"..\\windows\\system32\\cmd.exe",
	}
	for _, m := range malicious {
		_, err := checkZipSlip(dest, m)
		if err == nil {
			t.Errorf("expected checkZipSlip to fail for malicious path %q", m)
		}
	}
}

func TestFormatSize(t *testing.T) {
	if s := formatSize(500); s != "500 B" {
		t.Errorf("expected '500 B', got %q", s)
	}
	if s := formatSize(2048); s != "2.0 KB" {
		t.Errorf("expected '2.0 KB', got %q", s)
	}
	if s := formatSize(1048576 * 5); s != "5.0 MB" {
		t.Errorf("expected '5.0 MB', got %q", s)
	}
}

func TestMatchArchive(t *testing.T) {
	archives := []ArchiveItem{
		{
			FullPath: "/downloads/a.rar",
			Dir:      "/downloads",
			FileName: "a.rar",
			BaseName: "a",
			Format:   "RAR",
			Size:     1024,
			ModTime:  time.Now(),
		},
		{
			FullPath: "/downloads/project.zip",
			Dir:      "/downloads",
			FileName: "project.zip",
			BaseName: "project",
			Format:   "ZIP",
			Size:     2048,
			ModTime:  time.Now(),
		},
	}

	candidateDirs := []string{"/downloads"}

	// Match by index
	item, err := MatchArchive("1", archives, candidateDirs)
	if err != nil || item.FileName != "a.rar" {
		t.Fatalf("expected a.rar for index 1, got %v, err=%v", item, err)
	}

	item, err = MatchArchive("2", archives, candidateDirs)
	if err != nil || item.FileName != "project.zip" {
		t.Fatalf("expected project.zip for index 2, got %v, err=%v", item, err)
	}

	// Match by exact filename
	item, err = MatchArchive("a.rar", archives, candidateDirs)
	if err != nil || item.FileName != "a.rar" {
		t.Fatalf("expected a.rar, got %v, err=%v", item, err)
	}

	// Match by case-insensitive filename
	item, err = MatchArchive("A.RAR", archives, candidateDirs)
	if err != nil || item.FileName != "a.rar" {
		t.Fatalf("expected a.rar, got %v, err=%v", item, err)
	}

	// Match by base name
	item, err = MatchArchive("project", archives, candidateDirs)
	if err != nil || item.FileName != "project.zip" {
		t.Fatalf("expected project.zip, got %v, err=%v", item, err)
	}

	// Invalid input
	_, err = MatchArchive("99", archives, candidateDirs)
	if err == nil {
		t.Fatalf("expected error for invalid index 99")
	}

	_, err = MatchArchive("nonexistent.7z", archives, candidateDirs)
	if err == nil {
		t.Fatalf("expected error for nonexistent archive")
	}
}

func TestExtractRealArchives(t *testing.T) {
	// Test extracting existing sample.zip and sample.7z and test_full.rar
	zipItem := &ArchiveItem{
		FullPath: "/tmp/test_archives/sample.zip",
		Dir:      "/tmp/test_archives",
		FileName: "sample.zip",
		BaseName: "sample_unit",
		Format:   "ZIP",
	}
	outDir, count, err := ExtractArchive(zipItem, "/tmp/test_archives", "", false)
	if err != nil || count != 2 {
		t.Errorf("ExtractArchive zip failed: err=%v, count=%d", err, count)
	}
	if filepath.Base(outDir) != "sample_unit" {
		t.Errorf("unexpected outDir: %s", outDir)
	}
}
