package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/muslu/cemircol-go/cemircol"
)

func main() {
	// Sözlüğü yükle
	var emailList []string
	dictFile, err := os.Open("emails.dict")
	if err != nil {
		fmt.Println("❌ Sözlük dosyası bulunamadı. Önce parser'ı çalıştırın.")
		return
	}
	json.NewDecoder(dictFile).Decode(&emailList)
	dictFile.Close()

	statusNames := []string{"SENT", "DEFERRED", "BOUNCED", "EXPIRED"}

	fmt.Println("🔎 CemirCol Log Sorgulama Başlatılıyor...")
	
	rTime, _ := cemircol.NewReader("logs_time.cemir")
	rRcpt, _ := cemircol.NewReader("logs_rcpt.cemir")
	rStat, _ := cemircol.NewReader("logs_status.cemir")
	defer rTime.Close()
	defer rRcpt.Close()
	defer rStat.Close()

	times, _ := rTime.QueryInt64("time")
	rcpts, _ := rRcpt.QueryInt64("rcpt")
	stats, _ := rStat.QueryInt64("status")

	fmt.Printf("📊 Toplam %d kayıt bulundu.\n\n", len(stats))

	// Örnek sorgu: Son 10 hata (bounced)
	fmt.Println("⚠️ Son 10 Hatalı (BOUNCED) Gönderim:")
	count := 0
	for i := len(stats) - 1; i >= 0 && count < 10; i-- {
		if stats[i] == 2 { // 2 = bounced
			email := emailList[rcpts[i]]
			t := time.Unix(times[i], 0).Format("Jan 2 15:04:05")
			fmt.Printf("   [%s] %-10s -> %s\n", t, statusNames[stats[i]], email)
			count++
		}
	}

	// İstatistikler
	fmt.Println("\n📈 Genel İstatistikler:")
	statCounts := make(map[int64]int)
	for _, s := range stats {
		statCounts[s]++
	}
	for i, name := range statusNames {
		fmt.Printf("   %-10s: %d\n", name, statCounts[int64(i)])
	}
}
