package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bodgit/sevenzip"
)

func extract7z(src, destDir, password string, verbose bool) (int, error) {
	var r *sevenzip.ReadCloser
	var err error

	if password != "" {
		r, err = sevenzip.OpenReaderWithPassword(src, password)
	} else {
		r, err = sevenzip.OpenReader(src)
	}

	if err != nil {
		return 0, fmt.Errorf("7Z arşivi açılamadı: %w", err)
	}
	defer r.Close()

	count := 0
	for _, f := range r.File {
		target, err := checkZipSlip(destDir, f.Name)
		if err != nil {
			fmt.Printf("  [UYARI] Atlandı: %v\n", err)
			continue
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return count, err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return count, err
		}

		rc, err := f.Open()
		if err != nil {
			return count, fmt.Errorf("dosya okunamadı (%s): %w", f.Name, err)
		}

		mode := f.Mode()
		if mode == 0 {
			mode = 0644
		}
		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			rc.Close()
			return count, fmt.Errorf("dosya oluşturulamadı (%s): %w", target, err)
		}

		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()
		if err != nil {
			return count, fmt.Errorf("dosya çıkartılırken hata (%s): %w", f.Name, err)
		}

		if !f.Modified.IsZero() {
			_ = os.Chtimes(target, f.Modified, f.Modified)
		}

		count++
		if verbose {
			fmt.Printf("  -> [OK] %s (%s)\n", f.Name, formatSize(int64(f.UncompressedSize)))
		}
	}

	return count, nil
}
