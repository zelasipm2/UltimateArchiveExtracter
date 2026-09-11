package main

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ulikunitz/xz"
)

func extractTar(src, destDir, format string, verbose bool) (int, error) {
	f, err := os.Open(src)
	if err != nil {
		return 0, fmt.Errorf("dosya açılamadı: %w", err)
	}
	defer f.Close()

	var stream io.Reader = f

	switch format {
	case "TAR.GZ":
		gz, err := gzip.NewReader(f)
		if err != nil {
			return 0, fmt.Errorf("GZIP akışı açılamadı: %w", err)
		}
		defer gz.Close()
		stream = gz
	case "TAR.BZ2":
		stream = bzip2.NewReader(f)
	case "TAR.XZ":
		xr, err := xz.NewReader(f)
		if err != nil {
			return 0, fmt.Errorf("XZ akışı açılamadı: %w", err)
		}
		stream = xr
	}

	tr := tar.NewReader(stream)
	count := 0

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, fmt.Errorf("TAR okunurken hata: %w", err)
		}

		target, err := checkZipSlip(destDir, header.Name)
		if err != nil {
			fmt.Printf("  [UYARI] Atlandı: %v\n", err)
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return count, err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return count, err
			}

			mode := header.FileInfo().Mode()
			if mode == 0 {
				mode = 0644
			}
			out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
			if err != nil {
				return count, fmt.Errorf("dosya oluşturulamadı (%s): %w", target, err)
			}

			_, err = io.Copy(out, tr)
			out.Close()
			if err != nil {
				return count, fmt.Errorf("dosya çıkartılırken hata (%s): %w", header.Name, err)
			}

			if !header.ModTime.IsZero() {
				_ = os.Chtimes(target, header.ModTime, header.ModTime)
			}

			count++
			if verbose {
				fmt.Printf("  -> [OK] %s (%s)\n", header.Name, formatSize(header.Size))
			}
		}
	}

	return count, nil
}

func extractSingleCompressed(src, destDir, format, baseName string) (int, error) {
	f, err := os.Open(src)
	if err != nil {
		return 0, fmt.Errorf("dosya açılamadı: %w", err)
	}
	defer f.Close()

	var reader io.Reader
	switch format {
	case "GZ":
		gz, err := gzip.NewReader(f)
		if err != nil {
			return 0, err
		}
		defer gz.Close()
		reader = gz
	case "BZ2":
		reader = bzip2.NewReader(f)
	case "XZ":
		xr, err := xz.NewReader(f)
		if err != nil {
			return 0, err
		}
		reader = xr
	default:
		return 0, fmt.Errorf("bilinmeyen sıkıştırma: %s", format)
	}

	targetPath := filepath.Join(destDir, baseName)
	out, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	n, err := io.Copy(out, reader)
	if err != nil {
		return 0, err
	}

	fmt.Printf("  -> [OK] %s (%s)\n", baseName, formatSize(n))
	return 1, nil
}
