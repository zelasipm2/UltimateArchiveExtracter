#!/usr/bin/env bash
set -e

echo "=== Building UltimateArchiveExtracter ==="

# Build Windows 64-bit standalone executable
echo "[1/2] Windows amd64 .exe derleniyor..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o UltimateArchiveExtracter.exe .
cp UltimateArchiveExtracter.exe X.exe
echo "      -> UltimateArchiveExtracter.exe & X.exe oluşturuldu."

# Build Linux 64-bit binary
echo "[2/2] Linux amd64 ikili dosyası derleniyor..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o UltimateArchiveExtracter .
echo "      -> UltimateArchiveExtracter oluşturuldu."

echo "=== Derleme Başarıyla Tamamlandı! ==="
ls -lh UltimateArchiveExtracter.exe X.exe UltimateArchiveExtracter
