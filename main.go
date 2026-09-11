package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

const version = "1.0.0"

func printBanner() {
	fmt.Println("================================================================")
	fmt.Println("             ULTIMATE ARCHIVE EXTRACTOR (CLI)                   ")
	fmt.Println("      Evrensel Arşiv Çıkarıcı - ZIP / RAR / 7Z / TAR / GZ       ")
	fmt.Println("================================================================")
}

func printHelp() {
	printBanner()
	fmt.Println(`Kullanım Biçimleri:

1) İnteraktif Mod (Tavsiye Edilen):
   X.exe
   (Klasörleri otomatik tarar, arşivleri listeler ve seçim yapmanızı ister)

2) Doğrudan Komut Satırından Çıkarma:
   X.exe a.rar
   X.exe a.zip
   X.exe "C:\Users\Admin\Downloads\a.7z"
   (Belirtilen arşivi doğrudan kendi adındaki klasöre çıkartır)

3) Parametreler ile Kullanım:
   X.exe -d <klasör>            Belirtilen klasördeki arşivleri tarar
   X.exe -o <hedef_klasör>      Çıkartılacak üst klasörü belirler
   X.exe -p <şifre>             Şifreli arşivler için parola
   X.exe -l                     Yalnızca arşivleri listeler ve çıkar
   X.exe -h, --help             Bu yardım menüsünü gösterir

Örnekler:
   X.exe
   X.exe a.rar
   X.exe -d "C:\Users\Admin\Downloads"
   X.exe -o "D:\Cikarilanlar" a.rar`)
}

func printArchivesList(archives []ArchiveItem) {
	fmt.Printf("\n[*] Bulunan Arşivler (%d adet):\n", len(archives))
	fmt.Println("----------------------------------------------------------------")
	for i, item := range archives {
		fmt.Printf("  [%d] %-20s (%-8s) [%s]\n", i+1, item.FileName, item.Format, formatSize(item.Size))
		fmt.Printf("      -> Yol: %s\n", item.FullPath)
	}
	fmt.Println("----------------------------------------------------------------")
}

func main() {
	initConsole()

	dirFlag := flag.String("d", "", "Taranacak özel klasör yolu")
	outFlag := flag.String("o", "", "Özel çıkarma hedef klasörü")
	passFlag := flag.String("p", "", "Arşiv parolası")
	listOnly := flag.Bool("l", false, "Yalnızca arşivleri listele")
	helpFlag := flag.Bool("h", false, "Yardım mesajı")
	flag.BoolVar(helpFlag, "help", false, "Yardım mesajı")

	flag.Usage = printHelp
	flag.Parse()

	if *helpFlag {
		printHelp()
		return
	}

	args := flag.Args()
	candidateDirs := GetCandidateDirectories(*dirFlag)

	// Durum 1: Komut satırından doğrudan arşiv ismi/yolu verilmişse
	// Örnek: X.exe a.rar veya X.exe a.zip
	if len(args) > 0 {
		printBanner()
		targetInput := args[0]
		archives, _ := ScanArchives(candidateDirs)

		matchedItem, err := MatchArchive(targetInput, archives, candidateDirs)
		if err != nil {
			fmt.Printf("[-] HATA: %v\n", err)
			os.Exit(1)
		}

		destDir, count, err := ExtractArchive(matchedItem, *outFlag, *passFlag, true)
		if err != nil {
			fmt.Printf("[-] HATA: Çıkarma işlemi başarısız: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("----------------------------------------------------------------")
		fmt.Printf("[✓] BAŞARILI: '%s' arşivi çıkartıldı!\n", matchedItem.FileName)
		fmt.Printf("[✓] Hedef Klasör: %s\n", destDir)
		fmt.Printf("[✓] Toplam Çıkartılan Dosya: %d\n", count)
		return
	}

	// Durum 2: İnteraktif Mod (X.exe doğrudan çalıştırıldı)
	printBanner()

	fmt.Println("[*] Taranan Konumlar:")
	for _, d := range candidateDirs {
		fmt.Printf("  • %s\n", d)
	}

	archives, err := ScanArchives(candidateDirs)
	if err != nil {
		fmt.Printf("[-] Tarama hatası: %v\n", err)
	}

	if *listOnly {
		if len(archives) == 0 {
			fmt.Println("\n[-] Herhangi bir arşiv bulunamadı.")
		} else {
			printArchivesList(archives)
		}
		return
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		// Her turda listeyi göster
		if len(archives) == 0 {
			fmt.Println("\n[-] Taranan konumlarda herhangi bir arşiv (.zip, .rar, .7z, .tar, .gz) bulunamadı.")
			fmt.Println("[*] Çıkartmak istediğiniz arşivin tam yolunu yazabilir veya çıkmak için 'q' girebilirsiniz:")
		} else {
			printArchivesList(archives)
			fmt.Print("\nÇıkarmak istediğiniz arşivin adını veya numarasını yazın (örn: a.rar veya 1) [Çıkış: q]:\n> ")
		}

		inputLine, err := reader.ReadString('\n')
		if err != nil {
			// EOF or stdin closed
			break
		}

		input := strings.TrimSpace(inputLine)
		if input == "" {
			continue
		}
		lower := strings.ToLower(input)
		if lower == "q" || lower == "exit" || lower == "quit" || lower == "cikis" || lower == "çıkış" {
			fmt.Println("[*] Çıkış yapıldı.")
			break
		}

		matchedItem, err := MatchArchive(input, archives, candidateDirs)
		if err != nil {
			fmt.Printf("[-] %v. Lütfen tekrar deneyin.\n\n", err)
			continue
		}

		// Şifreli arşiv kontrolü (eğer şifre bayrağı girilmemişse ve arşiv açılırken şifre gerekirse)
		destDir, count, err := ExtractArchive(matchedItem, *outFlag, *passFlag, true)
		if err != nil {
			// Eğer şifre hatası olabilirse kullanıcıdan şifre iste
			if strings.Contains(strings.ToLower(err.Error()), "password") || strings.Contains(strings.ToLower(err.Error()), "şifre") || strings.Contains(strings.ToLower(err.Error()), "encrypted") {
				fmt.Print("[!] Arşiv şifreli görünüyor. Lütfen parolayı girin: ")
				passLine, _ := reader.ReadString('\n')
				pass := strings.TrimSpace(passLine)
				if pass != "" {
					destDir, count, err = ExtractArchive(matchedItem, *outFlag, pass, true)
				}
			}
		}

		if err != nil {
			fmt.Printf("[-] HATA: Çıkarma işlemi tamamlanamadı: %v\n\n", err)
		} else {
			fmt.Println("----------------------------------------------------------------")
			fmt.Printf("[✓] BAŞARILI: '%s' arşivi çıkartıldı!\n", matchedItem.FileName)
			fmt.Printf("[✓] Hedef Klasör: %s\n", destDir)
			fmt.Printf("[✓] Toplam Çıkartılan Dosya: %d\n", count)
			fmt.Println("----------------------------------------------------------------")
		}

		// Yeniden tarama yap (dosyalar güncellenmiş olabilir)
		archives, _ = ScanArchives(candidateDirs)

		fmt.Print("Başka bir arşiv çıkarmak istiyor musunuz? (Devam etmek için Enter, çıkmak için 'q'): ")
		nextInput, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		nextClean := strings.ToLower(strings.TrimSpace(nextInput))
		if nextClean == "q" || nextClean == "exit" || nextClean == "quit" || nextClean == "hayır" || nextClean == "no" || nextClean == "n" {
			fmt.Println("[*] İşlem tamamlandı. Görüşmek üzere!")
			break
		}
	}
}
