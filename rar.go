package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nwaples/rardecode/v2"
)

func extractRar(src, destDir, password string, verbose bool) (int, error) {
	var opts []rardecode.Option
	if password != "" {
		opts = append(opts, rardecode.Password(password))
	}

	r, err := rardecode.OpenReader(src, opts...)
	if err != nil {
		return 0, fmt.Errorf("RAR arşivi açılamadı: %w", err)
	}
	defer r.Close()

	count := 0
	for {
		header, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, fmt.Errorf("RAR okunurken hata: %w", err)
		}

		target, err := checkZipSlip(destDir, header.Name)
		if err != nil {
			fmt.Printf("  [UYARI] Atlandı: %v\n", err)
			continue
		}

		if header.IsDir {
			if err := os.MkdirAll(target, 0755); err != nil {
				return count, err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return count, err
		}

		mode := header.Mode()
		if mode == 0 {
			mode = 0644
		}
		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			return count, fmt.Errorf("dosya oluşturulamadı (%s): %w", target, err)
		}

		_, err = io.Copy(out, r)
		out.Close()
		if err != nil {
			return count, fmt.Errorf("dosya çıkartılırken hata (%s): %w", header.Name, err)
		}

		if !header.ModificationTime.IsZero() {
			_ = os.Chtimes(target, header.ModificationTime, header.ModificationTime)
		}

		count++
		if verbose {
			fmt.Printf("  -> [OK] %s (%s)\n", header.Name, formatSize(header.UnPackedSize))
		}
	}

	return count, nil
}
