package main

import (
	"fmt"
	"log"
	"time"
	"github.com/muslu/cemircol-go/cemircol"
)

func main() {
	const numRows = 10_000_000
	const filename = "data.cemir"

	fmt.Printf("🚀 10 Milyon satırlık veri oluşturuluyor (%d MB)...\n", numRows*8/1024/1024)
	
	data := make([]float64, numRows)
	for i := 0; i < numRows; i++ {
		data[i] = float64(i) * 0.1
	}

	start := time.Now()
	err := cemircol.WriteFloat64(filename, "val", data)
	if err != nil {
		log.Fatalf("Yazma hatası: %v", err)
	}
	writeTime := time.Since(start)
	fmt.Printf("✅ Yazma tamamlandı: %v\n", writeTime)

	fmt.Println("📖 Veri okunuyor (mmap + zero-copy)...")
	start = time.Now()
	reader, err := cemircol.NewReader(filename)
	if err != nil {
		log.Fatalf("Okuma hatası: %v", err)
	}
	defer reader.Close()

	readData, err := reader.QueryFloat64("val")
	if err != nil {
		log.Fatalf("Sorgu hatası: %v", err)
	}
	readTime := time.Since(start)

	fmt.Printf("✅ Okuma tamamlandı: %v\n", readTime)
	fmt.Printf("📊 Satır sayısı: %d\n", reader.NumRows())
	fmt.Printf("🔍 İlk 3 değer: %v\n", readData[:3])
	fmt.Printf("🔍 Son 3 değer: %v\n", readData[len(readData)-3:])
	
	fmt.Printf("\n⚡ Performans Özeti:\n")
	fmt.Printf("- Yazma Hızı: %.2f M satır/sn\n", float64(numRows)/writeTime.Seconds()/1_000_000)
	fmt.Printf("- Okuma Hızı: %.2f M satır/sn\n", float64(numRows)/readTime.Seconds()/1_000_000)
}
