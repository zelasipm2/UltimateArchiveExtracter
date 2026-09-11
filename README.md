# UltimateArchiveExtracter (X.exe)

Windows CLI ortamlarında (özellikle **`system32/systemprofile`**, Windows Servisleri, PsExec, Görev Yöneticisi veya standart kullanıcı oturumlarında) çalışan, hiçbir ek bağımlılık (Python, WinRAR, 7-Zip vb.) gerektirmeyen **tamamen bağımsız (standalone)** arşiv çıkarma aracıdır.

---

## 🚀 Temel Özellikler

1. **Tamamen Bağımsız (Standalone .EXE):**
   - Tek bir `.exe` dosyasından ibarettir (`UltimateArchiveExtracter.exe` veya `X.exe`).
   - Sistemde Python, .NET Runtime, Visual C++ Redistributable, WinRAR veya 7-Zip kurulu olmasına **gerek yoktur**.
   - Doğrudan makine koduna (x86_64 PE) derlenmiştir.

2. **Geniş Arşiv Desteği:**
   - **RAR** (`.rar` - RAR4 ve RAR5 biçimleri)
   - **ZIP** (`.zip`)
   - **7-Zip** (`.7z`)
   - **TAR** (`.tar`, `.tar.gz`, `.tgz`, `.tar.bz2`, `.tbz2`, `.tar.xz`, `.txz`)
   - **Sıkıştırılmış Tekil Dosyalar** (`.gz`, `.bz2`, `.xz`)

3. **`system32/systemprofile` ve Akıllı Klasör Taraması:**
   - Windows'ta `SYSTEM` (`LocalSystem`) kullanıcısı altında çalışırken `%USERPROFILE%` yolu `C:\Windows\System32\config\systemprofile` olur ve genellikle bu konumda bir `Downloads` klasörü bulunmaz veya boştur.
   - UltimateArchiveExtracter otomatik olarak:
     - Mevcut çalışma dizinini (`.`)
     - `.exe` dosyasının bulunduğu klasörü
     - `systemprofile\Downloads` dizinini
     - Windows Kayıt Kütüğü'ndeki (Registry) `Downloads` shell klasörünü
     - **`C:\Users\*\Downloads` (Sistemdeki tüm kullanıcıların İndirilenler klasörlerini)**
     - `C:\Windows\Temp` ve masaüstü konumlarını
     tarar ve indirilen arşivleri anında bulur!

4. **Arşiv Adıyla Otomatik Klasöre Çıkarma:**
   - Bir arşiv seçildiğinde (örneğin `a.rar` veya `a.zip`), arşivin bulunduğu konumda arşiv adında bir klasör (`a\`) oluşturur ve tüm içeriği bu klasörün içine çıkartır.
   - İndirilenler klasörünün içine yüzlerce dosyanın dağılmasını engeller.

5. **Güvenlik (Zip Slip Koruması):**
   - Kötü niyetli hazırlanmış `..\..\` tarzı dizin aşma (path traversal) saldırılarını engeller.

---

## 💻 Kullanım Şekilleri

### 1. İnteraktif Kullanım (Sadece `X.exe` Yazıldığında)
Komut satırında doğrudan `X.exe` (veya `UltimateArchiveExtracter.exe`) çalıştırıldığında:

```cmd
C:\Windows\system32> X.exe
================================================================
             ULTIMATE ARCHIVE EXTRACTOR (CLI)                   
      Evrensel Arşiv Çıkarıcı - ZIP / RAR / 7Z / TAR / GZ       
================================================================
[*] Taranan Konumlar:
  • C:\Windows\System32
  • C:\Users\Admin\Downloads
  • C:\Windows\System32\config\systemprofile\Downloads

[*] Bulunan Arşivler (3 adet):
----------------------------------------------------------------
  [1] a.rar                (RAR     ) [4.5 MB]
      -> Yol: C:\Users\Admin\Downloads\a.rar
  [2] test.zip             (ZIP     ) [1.2 MB]
      -> Yol: C:\Users\Admin\Downloads\test.zip
  [3] backup.7z            (7Z      ) [150.0 MB]
      -> Yol: C:\Windows\System32\backup.7z
----------------------------------------------------------------

Çıkarmak istediğiniz arşivin adını veya numarasını yazın (örn: a.rar veya 1) [Çıkış: q]:
> a.rar

[*] Arşiv Açılıyor : a.rar
[*] Format         : RAR (4.5 MB)
[*] Hedef Klasör   : C:\Users\Admin\Downloads\a
----------------------------------------------------------------
  -> [OK] dosya1.txt (15.2 KB)
  -> [OK] resim.png (1.2 MB)
----------------------------------------------------------------
[✓] BAŞARILI: 'a.rar' arşivi çıkartıldı!
[✓] Hedef Klasör: C:\Users\Admin\Downloads\a
[✓] Toplam Çıkartılan Dosya: 2
```

> Kullanıcı isterse `a.rar`, isterse `a.zip`, isterse listenin sıra numarasını (`1`, `2` vb.), isterse uzantısız `a` yazabilir.

---

### 2. Doğrudan Komut Satırından Çıkarma (Parametre ile)
İnteraktif menüye girmeden doğrudan arşivi çıkartmak için:

```cmd
X.exe a.rar
```
veya
```cmd
X.exe a.zip
```
veya tam yol belirterek:
```cmd
X.exe "C:\Users\Admin\Downloads\proje.7z"
```

---

### 3. Ek Parametreler ve Seçenekler

```cmd
X.exe -h                      # Yardım menüsünü görüntüler
X.exe -l                      # Yalnızca bulunan arşivleri listeler
X.exe -d "D:\Arsivler"        # Belirli bir klasörü tarar
X.exe -o "D:\Cikarilanlar" a.rar  # Çıkarılacak üst dizini belirler
X.exe -p "gizlisifre" gizli.rar   # Şifreli arşivler için parola girer
```

---

## 🛠 Yeniden Derleme (Build)

Projeyi Linux veya Windows üzerinde derlemek için:

```bash
./build.sh
```

Veya elle:

```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o UltimateArchiveExtracter.exe .
cp UltimateArchiveExtracter.exe X.exe
```

Testleri çalıştırmak için:

```bash
go test -v ./...
```
