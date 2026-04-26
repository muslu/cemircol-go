package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"
)

type MailRecord struct {
	Timestamp int64  `parquet:"name=timestamp, type=INT64"`
	Recipient string `parquet:"name=recipient, type=BYTE_ARRAY, convertedtype=UTF8"`
	Status    string `parquet:"name=status, type=BYTE_ARRAY, convertedtype=UTF8"`
}

func main() {
	fmt.Println("🔎 Parquet Log Sorgulama Başlatılıyor...")

	fr, err := local.NewLocalFileReader("logs.parquet")
	if err != nil {
		log.Fatalf("Parquet dosyası bulunamadı: %v", err)
	}
	defer fr.Close()

	pr, err := reader.NewParquetReader(fr, new(MailRecord), 4)
	if err != nil {
		log.Fatalf("Parquet okuyucu hatası: %v", err)
	}
	defer pr.ReadStop()

	num := int(pr.GetNumRows())
	records := make([]MailRecord, num)
	if err := pr.Read(&records); err != nil {
		log.Fatalf("Okuma hatası: %v", err)
	}

	fmt.Printf("📊 Toplam %d kayıt bulundu.\n\n", len(records))

	fmt.Println("⚠️ Son 10 Hatalı (bounced) Gönderim:")
	count := 0
	for i := len(records) - 1; i >= 0 && count < 10; i-- {
		if records[i].Status == "bounced" {
			t := time.Unix(records[i].Timestamp, 0).Format("Jan 2 15:04:05")
			fmt.Printf("   [%s] BOUNCED    -> %s\n", t, records[i].Recipient)
			count++
		}
	}

	// İstatistikler
	fmt.Println("\n📈 Genel İstatistikler:")
	stats := make(map[string]int)
	for _, r := range records {
		stats[r.Status]++
	}
	for status, c := range stats {
		fmt.Printf("   %-10s: %d\n", strings.ToUpper(status), c)
	}
}
